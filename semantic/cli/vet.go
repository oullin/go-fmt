package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/oullin/go-fmt/semantic/config"
	"github.com/oullin/go-fmt/semantic/engine"
)

func resolveVetRoot(workRoot string, cfg config.Config) (string, error) {
	if !cfg.Rules.Vet.Enabled {
		return "", nil
	}

	return discoverVetRoot(workRoot)
}

func discoverVetRoot(workRoot string) (string, error) {
	goWork, err := goEnv(workRoot, "GOWORK")

	if err != nil {
		return "", err
	}

	if root, ok := existingGoRoot(goWork, "go.work"); ok {
		return root, nil
	}

	goMod, err := goEnv(workRoot, "GOMOD")

	if err != nil {
		return "", err
	}

	if root, ok := existingGoRoot(goMod, "go.mod"); ok {
		return root, nil
	}

	return "", nil
}

func goEnv(workRoot string, key string) (string, error) {
	cmd := exec.Command("go", "env", key)
	cmd.Dir = workRoot

	out, err := cmd.Output()

	if err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("resolve go %s: %s", key, strings.TrimSpace(string(exitErr.Stderr)))
		}

		return "", fmt.Errorf("resolve go %s: %w", key, err)
	}

	return strings.TrimSpace(string(out)), nil
}

func existingGoRoot(path string, filename string) (string, bool) {
	if path == "" || path == "off" {
		return "", false
	}

	if filepath.Clean(path) == filepath.Clean(os.DevNull) {
		return "", false
	}

	info, err := os.Stat(path)

	if err != nil || info.IsDir() || filepath.Base(path) != filename {
		return "", false
	}

	return filepath.Dir(path), true
}

func runGoVet(root string) *engine.ErrorResult {
	if strings.TrimSpace(root) == "" {
		return nil
	}

	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = root

	out, err := cmd.CombinedOutput()

	if err == nil {
		return nil
	}

	message := "automatic go vet ./... failed"
	trimmed := strings.TrimSpace(string(bytes.TrimSpace(out)))

	if trimmed != "" {
		message = fmt.Sprintf("%s:\n%s", message, trimmed)
	} else {
		message = fmt.Sprintf("%s: %v", message, err)
	}

	return &engine.ErrorResult{
		File:    root,
		Message: message,
	}
}
