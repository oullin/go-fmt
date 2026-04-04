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

	absRoot, err := filepath.Abs(root)

	if err != nil {
		return nil, fmt.Errorf("resolve git diff root: %w", err)
	}

	repoRoot, err := gitRepoRoot(absRoot)

	if err != nil {
		return nil, err
	}

	cmd := exec.Command("git", "-C", absRoot, "diff", "--name-only", "--diff-filter=ACMR", "HEAD")

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

		path := filepath.Join(repoRoot, filepath.FromSlash(string(line)))

		if filepath.Ext(path) != ".go" {
			continue
		}

		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		}
	}

	return files, nil
}

func gitRepoRoot(root string) (string, error) {
	cmd := exec.Command("git", "-C", root, "rev-parse", "--show-toplevel")

	output, err := cmd.Output()

	if err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("git rev-parse: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}

		return "", fmt.Errorf("git rev-parse: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
