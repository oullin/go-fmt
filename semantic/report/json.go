package report

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/oullin/go-fmt/semantic/engine"
)

type jsonReport struct {
	Result  string             `json:"result"`
	Files   int                `json:"files"`
	Changed int                `json:"changed"`
	Results []jsonFileResult   `json:"results,omitempty"`
	Errors  []jsonErrorMessage `json:"errors,omitempty"`
}

type jsonFileResult struct {
	File       string          `json:"file"`
	Applied    []string        `json:"applied,omitempty"`
	Violations []jsonViolation `json:"violations,omitempty"`
	Changed    bool            `json:"changed,omitempty"`
}

type jsonViolation struct {
	Rule    string `json:"rule"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

func RenderJSON(w io.Writer, cwd string, report engine.Report) error {
	return json.NewEncoder(w).Encode(toJSONReport(cwd, report))
}

func toJSONReport(cwd string, report engine.Report) jsonReport {
	out := jsonReport{
		Result:  report.Result,
		Files:   report.Files,
		Changed: report.Changed,
	}

	for _, result := range report.Results {
		rel := relativePath(cwd, result.File)

		if strings.TrimSpace(result.Error) != "" {
			out.Errors = append(out.Errors, jsonErrorMessage{
				File:    rel,
				Message: result.Error,
			})

			continue
		}

		item := jsonFileResult{
			File:    rel,
			Applied: result.Applied,
			Changed: result.Changed,
		}

		for _, violation := range result.Violations {
			item.Violations = append(item.Violations, jsonViolation{
				Rule:    violation.Rule,
				Line:    violation.Line,
				Message: violation.Message,
			})
		}

		if item.Changed || len(item.Violations) > 0 {
			out.Results = append(out.Results, item)
		}
	}

	return out
}
