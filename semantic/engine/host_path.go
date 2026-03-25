package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type HostPath string

const HostRootEnv = "HOST_PROJECT_PATH"

func (h HostPath) Resolve(workRoot string, positional []string) ([]string, error) {
	hostPath := string(h)

	if hostPath == "" {
		return positional, nil
	}

	if len(positional) > 0 {
		return nil, fmt.Errorf("--host-path cannot be used with positional paths")
	}

	if !filepath.IsAbs(hostPath) {
		return nil, fmt.Errorf("--host-path must be an absolute path")
	}

	hostRoot, ok := os.LookupEnv(HostRootEnv)

	if !ok || strings.TrimSpace(hostRoot) == "" {
		return nil, fmt.Errorf("--host-path requires %s to be set", HostRootEnv)
	}

	hostRootAbs, err := filepath.Abs(hostRoot)

	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", HostRootEnv, err)
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
		return nil, fmt.Errorf("--host-path must be within %s (%s)", HostRootEnv, hostRootAbs)
	}

	if rel == "." {
		return []string{workRoot}, nil
	}

	return []string{filepath.Join(workRoot, rel)}, nil
}
