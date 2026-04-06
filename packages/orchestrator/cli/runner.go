package cli

import (
	"fmt"
	"io"

	"github.com/oullin/go-fmt/packages/correctness"
	orchestratorreport "github.com/oullin/go-fmt/packages/orchestrator/report"
	"github.com/oullin/go-fmt/packages/semantic"
)

type ReportRenderer struct{}

type Outcome struct {
	Mode   Mode
	Result orchestratorreport.Combined
}

type Runner struct {
	Stdout   io.Writer
	Stderr   io.Writer
	Parser   Parser
	Planner  Planner
	Renderer ReportRenderer
}

func (ReportRenderer) Render(w io.Writer, format, cwd string, mode Mode, result orchestratorreport.Combined) error {
	return orchestratorreport.Render(w, format, cwd, mode.String(), result)
}

func (o Outcome) ExitCode() int {
	if o.Result.Correctness.ErrorCount() > 0 {
		return 1
	}

	if o.Mode == CheckMode {
		if o.Result.Semantic.Result == "pass" {
			return 0
		}

		return 1
	}

	if o.Result.Semantic.ErrorCount() > 0 {
		return 1
	}

	return 0
}

func NewRunner(stdout, stderr io.Writer) Runner {
	return Runner{
		Stdout:   stdout,
		Stderr:   stderr,
		Parser:   NewParser(stderr),
		Planner:  Planner{Environment: Environment{}, Semantic: semantic.Planner{}, Correctness: correctness.Planner{}},
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
