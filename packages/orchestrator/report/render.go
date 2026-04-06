package report

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/oullin/go-fmt/packages/correctness"
	semanticengine "github.com/oullin/go-fmt/packages/semantic/engine"
)

type Combined struct {
	Semantic    semanticengine.Report `json:"semantic"`
	Correctness correctness.Report    `json:"correctness"`
}

type jsonErrorMessage struct {
	File    string `json:"file"`
	Message string `json:"message"`
}

func Render(w io.Writer, format, cwd, mode string, report Combined) error {
	switch format {
	case "text":
		return RenderText(w, cwd, mode, report)
	case "json":
		return RenderJSON(w, cwd, report)
	case "agent":
		return RenderAgent(w, cwd, report)
	default:
		return errors.New("unsupported output format")
	}
}

func relativePath(root, path string) string {
	rel, err := filepath.Rel(root, path)

	if err != nil {
		return path
	}

	return rel
}

func combinedResult(report Combined) string {
	if report.Correctness.ErrorCount() > 0 || report.Semantic.ErrorCount() > 0 {
		return "fail"
	}

	return report.Semantic.Result
}

func correctnessStatus(report correctness.Report) string {
	switch {
	case report.Root == "":
		return "skipped"
	case report.ErrorCount() > 0:
		return "fail"
	default:
		return "pass"
	}
}
