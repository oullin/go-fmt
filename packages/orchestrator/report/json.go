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
	return json.NewEncoder(w).Encode(toJSONReport(cwd, report))
}

func toJSONReport(cwd string, report Combined) jsonReport {
	return jsonReport{
		Result:      combinedResult(report),
		Semantic:    toSemanticJSONReport(cwd, report),
		Correctness: toCorrectnessJSONReport(cwd, report),
	}
}

func toSemanticJSONReport(cwd string, report Combined) semanticJSONReport {
	out := semanticJSONReport{
		Result:  report.Semantic.Result,
		Files:   report.Semantic.Files,
		Changed: report.Semantic.Changed,
	}

	for _, result := range report.Semantic.Errors {
		out.Errors = append(out.Errors, jsonErrorMessage{
			File:    relativePath(cwd, result.File),
			Message: result.Message,
		})
	}

	for _, result := range report.Semantic.Results {
		rel := relativePath(cwd, result.File)

		if result.Error != "" {
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

func toCorrectnessJSONReport(cwd string, report Combined) correctnessJSONReport {
	out := correctnessJSONReport{
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
