package report

import (
	"encoding/json"
	"io"
)

type jsonReport struct {
	Result      string                `json:"result"`
	Semantic    semanticJSONReport    `json:"semantic"`
	Correctness correctnessJSONReport `json:"correctness"`
}

type semanticJSONReport struct {
	Result  string             `json:"result"`
	Files   int                `json:"files"`
	Changed int                `json:"changed"`
	Results []jsonFileResult   `json:"results,omitempty"`
	Errors  []jsonErrorMessage `json:"errors,omitempty"`
}

type correctnessJSONReport struct {
	Status string             `json:"status"`
	Errors []jsonErrorMessage `json:"errors,omitempty"`
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

func RenderJSON(w io.Writer, cwd string, report Combined) error {
	return json.NewEncoder(w).Encode(toJSONReport(projectReport(cwd, report)))
}

func toJSONReport(report projectedReport) jsonReport {
	return jsonReport{
		Result:      report.Result,
		Semantic:    toSemanticJSONReport(report.Semantic),
		Correctness: toCorrectnessJSONReport(report.Correctness),
	}
}

func toSemanticJSONReport(report projectedSemanticReport) semanticJSONReport {
	out := semanticJSONReport{
		Result:  report.Result,
		Files:   report.Files,
		Changed: report.Changed,
	}

	out.Errors = append(out.Errors, report.Errors...)

	for _, result := range report.Results {
		if result.Error != "" {
			continue
		}

		item := jsonFileResult{
			File:       result.File,
			Applied:    result.Applied,
			Changed:    result.Changed,
			Violations: result.Violations,
		}

		if item.Changed || len(item.Violations) > 0 {
			out.Results = append(out.Results, item)
		}
	}

	return out
}

func toCorrectnessJSONReport(report projectedCorrectnessReport) correctnessJSONReport {
	return correctnessJSONReport{
		Status: report.Status,
		Errors: report.Errors,
	}
}
