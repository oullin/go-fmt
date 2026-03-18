package engine

import "github.com/gocanto/go-cs-fixer/internal/rules"

type FileResult struct {
	File       string            `json:"file"`
	Applied    []string          `json:"applied,omitempty"`
	Violations []rules.Violation `json:"violations,omitempty"`
	Diff       string            `json:"diff,omitempty"`
	Error      string            `json:"error,omitempty"`
	Changed    bool              `json:"changed,omitempty"`
}

type Report struct {
	Result  string       `json:"result"`
	Files   int          `json:"files"`
	Changed int          `json:"changed"`
	Results []FileResult `json:"results,omitempty"`
}

func (r Report) ViolationCount() int {
	total := 0

	for _, result := range r.Results {
		total += len(result.Violations)
	}

	return total
}

func (r Report) ErrorCount() int {
	total := 0

	for _, result := range r.Results {
		if result.Error != "" {
			total++
		}
	}

	return total
}
