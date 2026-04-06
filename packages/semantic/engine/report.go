package engine

import "github.com/oullin/go-fmt/packages/semantic/rules"

type FileResult struct {
	File       string            `json:"file"`
	Applied    []string          `json:"applied,omitempty"`
	Violations []rules.Violation `json:"violations,omitempty"`
	Diff       string            `json:"diff,omitempty"`
	Error      string            `json:"error,omitempty"`
	Changed    bool              `json:"changed,omitempty"`
}

type ErrorResult struct {
	File    string `json:"file,omitempty"`
	Message string `json:"message"`
}

type Report struct {
	Result  string        `json:"result"`
	Files   int           `json:"files"`
	Changed int           `json:"changed"`
	Results []FileResult  `json:"results,omitempty"`
	Errors  []ErrorResult `json:"errors,omitempty"`
}

func (r Report) ViolationCount() int {
	total := 0

	for _, result := range r.Results {
		total += len(result.Violations)
	}

	return total
}

func (r Report) ErrorCount() int {
	total := len(r.Errors)

	for _, result := range r.Results {
		if result.Error != "" {
			total++
		}
	}

	return total
}

func (r Report) AllErrors() []ErrorResult {
	out := append([]ErrorResult(nil), r.Errors...)

	for _, result := range r.Results {
		if result.Error == "" {
			continue
		}

		out = append(out, ErrorResult{
			File:    result.File,
			Message: result.Error,
		})
	}

	return out
}
