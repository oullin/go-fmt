package semantic

import (
	"fmt"

	"github.com/oullin/go-fmt/packages/semantic/config"
	"github.com/oullin/go-fmt/packages/semantic/engine"
	"github.com/oullin/go-fmt/packages/semantic/formatter"
	"github.com/oullin/go-fmt/packages/semantic/rules"
	"github.com/oullin/go-fmt/packages/semantic/rules/spacing"
)

type Mode string

type RuleSet struct{}

type FormatterSet struct{}

type Planner struct {
	RuleSet      RuleSet
	FormatterSet FormatterSet
}

type BuildOptions struct {
	Mode     Mode
	Config   config.Config
	RunPaths []string
}

type Plan struct {
	Mode     Mode
	RunPaths []string
	Engine   *engine.Engine
}

const (
	CheckMode  Mode = "check"
	FormatMode Mode = "format"
)

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

func (p Planner) Build(options BuildOptions) (Plan, error) {
	switch options.Mode {
	case CheckMode, FormatMode:
	default:
		return Plan{}, fmt.Errorf("unsupported mode %q", options.Mode)
	}

	return Plan{
		Mode:     options.Mode,
		RunPaths: options.RunPaths,
		Engine:   engine.New(options.Config, p.RuleSet.Build(options.Config), p.FormatterSet.Build(options.Config)),
	}, nil
}

func (p Plan) Execute() (engine.Report, error) {
	switch p.Mode {
	case CheckMode:
		return p.Engine.Check(p.RunPaths)
	case FormatMode:
		return p.Engine.Format(p.RunPaths)
	default:
		return engine.Report{}, fmt.Errorf("unsupported mode %q", p.Mode)
	}
}
