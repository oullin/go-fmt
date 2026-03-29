package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/oullin/go-fmt/semantic/engine"
)

func RenderText(w io.Writer, cwd, mode string, report engine.Report) error {
	if report.Files == 0 {
		if _, err := color.New(color.FgYellow).Fprintf(w, "  No Go files found.\n"); err != nil {
			return err
		}

		return nil
	}

	action := "Checked"

	if mode == "format" {
		action = "Formatted"
	}

	if _, err := color.New(color.FgGreen, color.Bold).Fprintf(w, "  %s %d file(s).\n\n", action, report.Files); err != nil {
		return err
	}

	for _, result := range report.Results {
		if len(result.Violations) == 0 && result.Error == "" && !result.Changed {
			continue
		}

		rel := relativePath(cwd, result.File)

		if _, err := color.New(color.FgCyan, color.Bold).Fprintf(w, "  %s\n", rel); err != nil {
			return err
		}

		if result.Error != "" {
			if _, err := color.New(color.FgRed).Fprintf(w, "    ! %s\n", result.Error); err != nil {
				return err
			}

			fmt.Fprintln(w)

			continue
		}

		for _, violation := range result.Violations {
			ruleColor := color.New(color.FgMagenta).Sprintf("[%s]", violation.Rule)

			if violation.Line > 0 {
				if _, err := fmt.Fprintf(w, "    %s line %d: %s\n", ruleColor, violation.Line, violation.Message); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprintf(w, "    %s %s\n", ruleColor, violation.Message); err != nil {
					return err
				}
			}
		}

		if result.Changed {
			verb := "would apply"

			if mode == "format" {
				verb = "applied"
			}

			if _, err := color.New(color.FgGreen).Fprintf(w, "    ✓ %s %s\n", verb, strings.Join(result.Applied, ", ")); err != nil {
				return err
			}
		}

		fmt.Fprintln(w)
	}

	summaryColor := color.New(color.Bold)

	if report.ErrorCount() > 0 {
		summaryColor.Add(color.FgRed)
	} else if report.ViolationCount() > 0 {
		summaryColor.Add(color.FgYellow)
	} else {
		summaryColor.Add(color.FgGreen)
	}

	_, err := summaryColor.Fprintf(w, "  Result: %s. %d changed, %d violation(s), %d error(s).\n", report.Result, report.Changed, report.ViolationCount(), report.ErrorCount())

	return err
}
