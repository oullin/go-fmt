package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/oullin/go-fmt/packages/correctness"
	"github.com/oullin/go-fmt/packages/orchestrator/config"
	orchestratorreport "github.com/oullin/go-fmt/packages/orchestrator/report"
	"github.com/oullin/go-fmt/packages/semantic"
)

type Environment struct{}

type Planner struct {
	Environment Environment
	Semantic    semantic.Planner
	Correctness correctness.Planner
}

type Plan struct {
	Mode         Mode
	OutputFormat string
	ReportRoot   string
	Semantic     semantic.Plan
	Correctness  correctness.Plan
}

func (Environment) WorkingDirectory() (string, error) {
	return os.Getwd()
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

	semanticPlan, err := p.Semantic.Build(semantic.BuildOptions{
		Mode:     semantic.Mode(options.Mode),
		Config:   cfg.SemanticConfig(),
		RunPaths: runPaths,
	})

	if err != nil {
		return Plan{}, err
	}

	correctnessPlan, err := p.Correctness.Build(correctness.BuildOptions{
		WorkRoot: workRoot,
		Config:   cfg.CorrectnessConfig(),
	})

	if err != nil {
		return Plan{}, err
	}

	return Plan{
		Mode:         options.Mode,
		OutputFormat: options.OutputFormat,
		ReportRoot:   reportRoot,
		Semantic:     semanticPlan,
		Correctness:  correctnessPlan,
	}, nil
}

func (p Plan) Execute() (orchestratorreport.Combined, error) {
	semanticReport, err := p.Semantic.Execute()

	if err != nil {
		return orchestratorreport.Combined{}, err
	}

	return orchestratorreport.Combined{
		Semantic:    semanticReport,
		Correctness: p.Correctness.Execute(),
	}, nil
}
