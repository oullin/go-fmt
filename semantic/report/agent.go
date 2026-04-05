package report

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/oullin/go-fmt/semantic/engine"
)

type agentReport struct {
	Result     string             `json:"result"`
	Summary    agentSummary       `json:"summary"`
	Changed    []agentChange      `json:"changed,omitempty"`
	Violations []agentViolation   `json:"violations,omitempty"`
	Errors     []jsonErrorMessage `json:"errors,omitempty"`
}

type agentSummary struct {
	Files      int `json:"files"`
	Changed    int `json:"changed"`
	Violations int `json:"violations"`
}

type agentChange struct {
	File  string   `json:"file"`
	Steps []string `json:"steps"`
}

type agentViolation struct {
	File    string `json:"file"`
	Rule    string `json:"rule"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

func RenderAgent(w io.Writer, cwd string, report engine.Report) error {
	encoder := json.NewEncoder(w)

	encoder.SetIndent("", "  ")

	return encoder.Encode(toAgentReport(cwd, report))
}

func toAgentReport(cwd string, report engine.Report) agentReport {
	out := agentReport{
		Result: report.Result,
		Summary: agentSummary{
			Files:      report.Files,
			Changed:    report.Changed,
			Violations: report.ViolationCount(),
		},
	}

	for _, result := range report.Errors {
		out.Errors = append(out.Errors, jsonErrorMessage{
			File:    relativePath(cwd, result.File),
			Message: result.Message,
		})
	}

	for _, result := range report.Results {
		rel := relativePath(cwd, result.File)

		if result.Changed {
			out.Changed = append(out.Changed, agentChange{
				File:  rel,
				Steps: result.Applied,
			})
		}

		for _, violation := range result.Violations {
			out.Violations = append(out.Violations, agentViolation{
				File:    rel,
				Rule:    violation.Rule,
				Line:    violation.Line,
				Message: violation.Message,
			})
		}

		if strings.TrimSpace(result.Error) != "" {
			out.Errors = append(out.Errors, jsonErrorMessage{
				File:    rel,
				Message: result.Error,
			})
		}
	}

	return out
}
