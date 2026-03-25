package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/oullin/go-fmt/semantic/engine"
)

func RenderText(w io.Writer, cwd, mode string, report engine.Report) error {
	if report.Files == 0 {
		if _, err := fmt.Fprintln(w, "No Go files found."); err != nil {
			return err
		}

		return nil
	}

	action := "Checked"

	if mode == "format" {
		action = "Formatted"
	}

	if _, err := fmt.Fprintf(w, "%s %d file(s).\n", action, report.Files); err != nil {
		return err
	}

	for _, result := range report.Results {
		rel := relativePath(cwd, result.File)

		if result.Error != "" {
			if _, err := fmt.Fprintf(w, "! %s %s\n", rel, result.Error); err != nil {
				return err
			}

			continue
		}

		for _, violation := range result.Violations {
			if violation.Line > 0 {
				if _, err := fmt.Fprintf(w, "~ %s:%d [%s] %s\n", rel, violation.Line, violation.Rule, violation.Message); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprintf(w, "~ %s [%s] %s\n", rel, violation.Rule, violation.Message); err != nil {
					return err
				}
			}
		}

		if result.Changed {
			verb := "would apply"

			if mode == "format" {
				verb = "applied"
			}

			if _, err := fmt.Fprintf(w, "  %s %s\n", verb, strings.Join(result.Applied, ", ")); err != nil {
				return err
			}
		}
	}

	_, err := fmt.Fprintf(w, "Result: %s. %d changed, %d violation(s), %d error(s).\n", report.Result, report.Changed, report.ViolationCount(), report.ErrorCount())

	return err
}
