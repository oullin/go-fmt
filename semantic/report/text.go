package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/oullin/go-fmt/semantic/engine"
)

func RenderText(w io.Writer, cwd, mode string, report engine.Report) {
	if report.Files == 0 {
		fmt.Fprintln(w, "No Go files found.")

		return
	}

	action := "Checked"

	if mode == "format" {
		action = "Formatted"
	}

	fmt.Fprintf(w, "%s %d file(s).\n", action, report.Files)

	for _, result := range report.Results {
		rel := relativePath(cwd, result.File)

		if result.Error != "" {
			fmt.Fprintf(w, "! %s %s\n", rel, result.Error)

			continue
		}

		for _, violation := range result.Violations {
			if violation.Line > 0 {
				fmt.Fprintf(w, "~ %s:%d [%s] %s\n", rel, violation.Line, violation.Rule, violation.Message)
			} else {
				fmt.Fprintf(w, "~ %s [%s] %s\n", rel, violation.Rule, violation.Message)
			}
		}

		if result.Changed {
			verb := "would apply"

			if mode == "format" {
				verb = "applied"
			}

			fmt.Fprintf(w, "  %s %s\n", verb, strings.Join(result.Applied, ", "))
		}
	}

	fmt.Fprintf(w, "Result: %s. %d changed, %d violation(s), %d error(s).\n", report.Result, report.Changed, report.ViolationCount(), report.ErrorCount())
}
