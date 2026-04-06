package report

import (
	"encoding/json"
	"io"
)

type agentReport struct {
	Result      string                 `json:"result"`
	Semantic    semanticAgentReport    `json:"semantic"`
	Correctness correctnessAgentReport `json:"correctness"`
}

type semanticAgentReport struct {
	Result     string             `json:"result"`
	Summary    agentSummary       `json:"summary"`
	Changed    []agentChange      `json:"changed,omitempty"`
	Violations []agentViolation   `json:"violations,omitempty"`
	Errors     []jsonErrorMessage `json:"errors,omitempty"`
}

type correctnessAgentReport struct {
	Status string             `json:"status"`
	Errors []jsonErrorMessage `json:"errors,omitempty"`
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

func RenderAgent(w io.Writer, cwd string, report Combined) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	return encoder.Encode(toAgentReport(cwd, report))
}

func toAgentReport(cwd string, report Combined) agentReport {
	return agentReport{
		Result:      combinedResult(report),
		Semantic:    toSemanticAgentReport(cwd, report),
		Correctness: toCorrectnessAgentReport(cwd, report),
	}
}

func toSemanticAgentReport(cwd string, report Combined) semanticAgentReport {
	out := semanticAgentReport{
		Result: report.Semantic.Result,
		Summary: agentSummary{
			Files:      report.Semantic.Files,
			Changed:    report.Semantic.Changed,
			Violations: report.Semantic.ViolationCount(),
		},
	}

	for _, result := range report.Semantic.Errors {
		out.Errors = append(out.Errors, jsonErrorMessage{
			File:    relativePath(cwd, result.File),
			Message: result.Message,
		})
	}

	for _, result := range report.Semantic.Results {
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

		if result.Error != "" {
			out.Errors = append(out.Errors, jsonErrorMessage{
				File:    rel,
				Message: result.Error,
			})
		}
	}

	return out
}

func toCorrectnessAgentReport(cwd string, report Combined) correctnessAgentReport {
	out := correctnessAgentReport{
		Status: correctnessStatus(report.Correctness),
	}

	for _, result := range report.Correctness.Errors {
		out.Errors = append(out.Errors, jsonErrorMessage{
			File:    relativePath(cwd, result.File),
			Message: result.Message,
		})
	}

	return out
}
