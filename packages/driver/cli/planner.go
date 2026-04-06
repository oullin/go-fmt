package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/oullin/go-fmt/packages/driver/config"
	driverreport "github.com/oullin/go-fmt/packages/driver/report"
	"github.com/oullin/go-fmt/packages/formatter"
	"github.com/oullin/go-fmt/packages/vet"
)

type Environment struct{}

type Planner struct {
	Environment Environment
	Formatter   formatter.Planner
	Vet         vet.Planner
}

type Plan struct {
	Mode         Mode
	OutputFormat string
	ReportRoot   string
	Formatter    formatter.Plan
	Vet          vet.Plan
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

	formatterPlan, err := p.Formatter.Build(formatter.BuildOptions{
		Mode:     formatter.Mode(options.Mode),
		Config:   cfg.FormatterConfig(),
		RunPaths: runPaths,
	})

	if err != nil {
		return Plan{}, err
	}

	vetPlan, err := p.Vet.Build(vet.BuildOptions{
		WorkRoot: workRoot,
		Config:   cfg.VetConfig(),
	})

	if err != nil {
		return Plan{}, err
	}

	return Plan{
		Mode:         options.Mode,
		OutputFormat: options.OutputFormat,
		ReportRoot:   reportRoot,
		Formatter:    formatterPlan,
		Vet:          vetPlan,
	}, nil
}

func (p Plan) Execute() (driverreport.Combined, error) {
	formatterReport, err := p.Formatter.Execute()

	if err != nil {
		return driverreport.Combined{}, err
	}

	return driverreport.Combined{
		Formatter: formatterReport,
		Vet:       p.Vet.Execute(),
	}, nil
}
