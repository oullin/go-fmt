package cli

import (
	"fmt"
	"io"

	driverreport "github.com/oullin/go-fmt/packages/driver/report"
	"github.com/oullin/go-fmt/packages/formatter"
	"github.com/oullin/go-fmt/packages/vet"
)

type ReportRenderer struct{}

type Outcome struct {
	Mode   Mode
	Result driverreport.Combined
}

type Runner struct {
	Stdout   io.Writer
	Stderr   io.Writer
	Parser   Parser
	Planner  Planner
	Renderer ReportRenderer
}

func (ReportRenderer) Render(w io.Writer, format, cwd string, mode Mode, result driverreport.Combined) error {
	return driverreport.Render(w, format, cwd, mode.String(), result)
}

func (o Outcome) ExitCode() int {
	if o.Result.Vet.ErrorCount() > 0 {
		return 1
	}

	if o.Mode == CheckMode {
		if o.Result.Formatter.Result == "pass" {
			return 0
		}

		return 1
	}

	if o.Result.Formatter.ErrorCount() > 0 {
		return 1
	}

	return 0
}

func NewRunner(stdout, stderr io.Writer) Runner {
	return Runner{
		Stdout:   stdout,
		Stderr:   stderr,
		Parser:   NewParser(stderr),
		Planner:  Planner{Environment: Environment{}, Formatter: formatter.Planner{}, Vet: vet.Planner{}},
		Renderer: ReportRenderer{},
	}
}

func (r Runner) Run(mode Mode, args []string) int {
	options, err := r.Parser.Parse(mode, args)

	if err != nil {
		return 1
	}

	plan, err := r.Planner.Build(options)

	if err != nil {
		r.writeError("%v\n", err)

		return 1
	}

	result, err := plan.Execute()

	if err != nil {
		r.writeError("%v\n", err)

		return 1
	}

	if err := r.Renderer.Render(r.Stdout, plan.OutputFormat, plan.ReportRoot, mode, result); err != nil {
		r.writeError("render report: %v\n", err)

		return 1
	}

	return Outcome{Mode: mode, Result: result}.ExitCode()
}

func (r Runner) writeError(format string, args ...any) {
	_, _ = fmt.Fprintf(r.Stderr, format, args...)
}
