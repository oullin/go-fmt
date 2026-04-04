package engine

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GitDiffGoFiles(root string) ([]string, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("resolve git diff root: empty root")
	}

	cmd := exec.Command("git", "-C", root, "diff", "--name-only", "--diff-filter=ACMR", "HEAD")

	output, err := cmd.Output()

	if err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("git diff: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}

		return nil, fmt.Errorf("git diff: %w", err)
	}

	lines := bytes.Split(bytes.TrimSpace(output), []byte{'\n'})
	files := make([]string, 0, len(lines))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		path := filepath.Join(root, string(line))

		if filepath.Ext(path) != ".go" {
			continue
		}

		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		}
	}

	return files, nil
}
