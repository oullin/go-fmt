package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/oullin/go-fmt/semantic/config"
	"github.com/oullin/go-fmt/semantic/engine"
	"github.com/oullin/go-fmt/semantic/formatter"
	"github.com/oullin/go-fmt/semantic/rules"
	"github.com/oullin/go-fmt/semantic/rules/callback_extraction"
	"github.com/oullin/go-fmt/semantic/rules/declaration_order"
	"github.com/oullin/go-fmt/semantic/rules/spacing"
	"github.com/oullin/go-fmt/semantic/rules/trimspace"
)

type Environment struct{}

type RuleSet struct{}

type FormatterSet struct{}

type Planner struct {
	Environment  Environment
	RuleSet      RuleSet
	FormatterSet FormatterSet
}

type Plan struct {
	Mode         Mode
	OutputFormat string
	ReportRoot   string
	Files        []string
	Engine       *engine.Engine
}

func (Environment) WorkingDirectory() (string, error) {
	return os.Getwd()
}

func (Environment) GitDiffFiles(root string) ([]string, error) {
	return engine.GitDiffGoFiles(root)
}

func (RuleSet) Build(cfg config.Config) []rules.Rule {
	var out []rules.Rule

	if cfg.Rules.DeclarationOrder.Enabled {
		out = append(out, declaration_order.New())
	}

	if cfg.Rules.CallbackExtraction.Enabled {
		out = append(out, callback_extraction.New())
	}

	if cfg.Rules.TrimSpace.Enabled {
		out = append(out, trimspace.New())
	}

	if cfg.Rules.Spacing.Enabled {
		out = append(out, spacing.New())
	}

	return out
}

func (FormatterSet) Build(cfg config.Config) []formatter.Formatter {
	var out []formatter.Formatter

	if cfg.Formatters.Gofmt {
		out = append(out, formatter.NewGofmt())
	}

	if cfg.Formatters.Goimports {
		out = append(out, formatter.NewGoimports())
	}

	return out
}

func (p Planner) Build(options Options) (Plan, error) {
	workRoot, err := p.Environment.WorkingDirectory()

	if err != nil {
		return Plan{}, fmt.Errorf("resolve cwd: %w", err)
	}

	reportRoot := workRoot

	if strings.TrimSpace(options.ReportRoot) != "" {
		reportRoot = options.ReportRoot
	}

	cfg, err := config.Load(reportRoot, options.ConfigPath)

	if err != nil {
		return Plan{}, err
	}

	runPaths, err := options.HostPaths.Resolve(workRoot, options.Positional)

	if err != nil {
		return Plan{}, err
	}

	files, err := engine.CollectGoFiles(runPaths, cfg)

	if err != nil {
		return Plan{}, err
	}

	if options.GitDiff {
		diffFiles, err := p.Environment.GitDiffFiles(reportRoot)

		if err != nil {
			return Plan{}, err
		}

		files = engine.FilterFiles(files, diffFiles)
	}

	return Plan{
		Mode:         options.Mode,
		OutputFormat: options.OutputFormat,
		ReportRoot:   reportRoot,
		Files:        files,
		Engine:       engine.New(cfg, p.RuleSet.Build(cfg), p.FormatterSet.Build(cfg)),
	}, nil
}

func (p Plan) Execute() (engine.Report, error) {
	switch p.Mode {
	case CheckMode:
		return p.Engine.CheckFiles(p.Files)
	case FormatMode:
		return p.Engine.FormatFiles(p.Files)
	default:
		return engine.Report{}, fmt.Errorf("unsupported mode %q", p.Mode)
	}
}
