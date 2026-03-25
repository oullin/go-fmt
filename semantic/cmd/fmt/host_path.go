package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const hostRootEnv = "HOST_PROJECT_PATH"

func resolveRunPaths(workRoot, hostPath string, positional []string) ([]string, error) {
	if hostPath == "" {
		return positional, nil
	}

	if len(positional) > 0 {
		return nil, fmt.Errorf("--host-path cannot be used with positional paths")
	}

	if !filepath.IsAbs(hostPath) {
		return nil, fmt.Errorf("--host-path must be an absolute path")
	}

	hostRoot, ok := os.LookupEnv(hostRootEnv)

	if !ok || strings.TrimSpace(hostRoot) == "" {
		return nil, fmt.Errorf("--host-path requires %s to be set", hostRootEnv)
	}

	hostRootAbs, err := filepath.Abs(hostRoot)

	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", hostRootEnv, err)
	}

	hostPathAbs, err := filepath.Abs(hostPath)

	if err != nil {
		return nil, fmt.Errorf("resolve --host-path: %w", err)
	}

	rel, err := filepath.Rel(hostRootAbs, hostPathAbs)

	if err != nil {
		return nil, fmt.Errorf("map --host-path: %w", err)
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return nil, fmt.Errorf("--host-path must be within %s (%s)", hostRootEnv, hostRootAbs)
	}

	if rel == "." {
		return []string{workRoot}, nil
	}

	return []string{filepath.Join(workRoot, rel)}, nil
}
