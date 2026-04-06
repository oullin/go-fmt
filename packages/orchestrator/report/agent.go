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

	return encoder.Encode(toAgentReport(projectReport(cwd, report)))
}

func toAgentReport(report projectedReport) agentReport {
	return agentReport{
		Result:      report.Result,
		Semantic:    toSemanticAgentReport(report.Semantic),
		Correctness: toCorrectnessAgentReport(report.Correctness),
	}
}

func toSemanticAgentReport(report projectedSemanticReport) semanticAgentReport {
	out := semanticAgentReport{
		Result: report.Result,
		Summary: agentSummary{
			Files:      report.Files,
			Changed:    report.Changed,
			Violations: report.Violations,
		},
	}

	out.Errors = append(out.Errors, report.Errors...)

	for _, result := range report.Results {
		if result.Changed {
			out.Changed = append(out.Changed, agentChange{
				File:  result.File,
				Steps: result.Applied,
			})
		}

		for _, violation := range result.Violations {
			out.Violations = append(out.Violations, agentViolation{
				File:    result.File,
				Rule:    violation.Rule,
				Line:    violation.Line,
				Message: violation.Message,
			})
		}
	}

	return out
}

func toCorrectnessAgentReport(report projectedCorrectnessReport) correctnessAgentReport {
	return correctnessAgentReport{
		Status: report.Status,
		Errors: report.Errors,
	}
}
