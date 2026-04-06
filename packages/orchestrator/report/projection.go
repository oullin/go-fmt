package report

type projectedReport struct {
	Result      string
	Semantic    projectedSemanticReport
	Correctness projectedCorrectnessReport
}

type projectedSemanticReport struct {
	Result     string
	Files      int
	Changed    int
	Violations int
	Results    []projectedFileResult
	Errors     []jsonErrorMessage
}

type projectedCorrectnessReport struct {
	Status string
	Errors []jsonErrorMessage
}

type projectedFileResult struct {
	File       string
	Applied    []string
	Violations []jsonViolation
	Changed    bool
	Error      string
}

func projectReport(cwd string, report Combined) projectedReport {
	return projectedReport{
		Result:      combinedResult(report),
		Semantic:    projectSemanticReport(cwd, report),
		Correctness: projectCorrectnessReport(cwd, report),
	}
}

func projectSemanticReport(cwd string, report Combined) projectedSemanticReport {
	out := projectedSemanticReport{
		Result:     report.Semantic.Result,
		Files:      report.Semantic.Files,
		Changed:    report.Semantic.Changed,
		Violations: report.Semantic.ViolationCount(),
	}

	for _, result := range report.Semantic.Errors {
		out.Errors = append(out.Errors, jsonErrorMessage{
			File:    relativePath(cwd, result.File),
			Message: result.Message,
		})
	}

	for _, result := range report.Semantic.Results {
		item := projectedFileResult{
			File:    relativePath(cwd, result.File),
			Applied: result.Applied,
			Changed: result.Changed,
			Error:   result.Error,
		}

		for _, violation := range result.Violations {
			item.Violations = append(item.Violations, jsonViolation{
				Rule:    violation.Rule,
				Line:    violation.Line,
				Message: violation.Message,
			})
		}

		if item.Error != "" {
			out.Errors = append(out.Errors, jsonErrorMessage{
				File:    item.File,
				Message: item.Error,
			})
		}

		out.Results = append(out.Results, item)
	}

	return out
}

func projectCorrectnessReport(cwd string, report Combined) projectedCorrectnessReport {
	out := projectedCorrectnessReport{
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
