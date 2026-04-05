package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/oullin/go-fmt/semantic/config"
	"github.com/oullin/go-fmt/semantic/engine"
	"github.com/oullin/go-fmt/semantic/formatter"
	"github.com/oullin/go-fmt/semantic/rules"
	"github.com/oullin/go-fmt/semantic/rules/spacing"
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
	RunPaths     []string
	Engine       *engine.Engine
	VetRoot      string
}

func (Environment) WorkingDirectory() (string, error) {
	return os.Getwd()
}

func (RuleSet) Build(cfg config.Config) []rules.Rule {
	var out []rules.Rule

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

	runPaths, err := options.HostPath.Resolve(workRoot, options.Positional)

	if err != nil {
		return Plan{}, err
	}

	vetRoot, err := resolveVetRoot(workRoot, cfg)

	if err != nil {
		return Plan{}, err
	}

	return Plan{
		Mode:         options.Mode,
		OutputFormat: options.OutputFormat,
		ReportRoot:   reportRoot,
		RunPaths:     runPaths,
		Engine:       engine.New(cfg, p.RuleSet.Build(cfg), p.FormatterSet.Build(cfg)),
		VetRoot:      vetRoot,
	}, nil
}

func (p Plan) Execute() (engine.Report, error) {
	var (
		result engine.Report
		err    error
	)

	switch p.Mode {
	case CheckMode:
		result, err = p.Engine.Check(p.RunPaths)
	case FormatMode:
		result, err = p.Engine.Format(p.RunPaths)
	default:
		return engine.Report{}, fmt.Errorf("unsupported mode %q", p.Mode)
	}

	if err != nil {
		return engine.Report{}, err
	}

	if vetError := runGoVet(p.VetRoot); vetError != nil {
		result.Errors = append(result.Errors, *vetError)
		result.Result = "fail"
	}

	return result, nil
}
