package report

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/oullin/go-fmt/engine"
)

func Render(w io.Writer, format, cwd, mode string, report engine.Report) error {
	switch format {
	case "text":
		RenderText(w, cwd, mode, report)

		return nil
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
