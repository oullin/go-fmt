package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const HostRootEnv = "HOST_PROJECT_PATH"

type HostPath string

type HostPaths []HostPath

func (h HostPath) Resolve(workRoot string, positional []string) ([]string, error) {
	return HostPaths{h}.Resolve(workRoot, positional)
}

func (h HostPaths) Resolve(workRoot string, positional []string) ([]string, error) {
	if len(h) == 0 {
		return positional, nil
	}

	if len(positional) > 0 {
		return nil, fmt.Errorf("--host-path cannot be used with positional paths")
	}

	hostRoot, ok := os.LookupEnv(HostRootEnv)

	if !ok || strings.TrimSpace(hostRoot) == "" {
		return nil, fmt.Errorf("--host-path requires %s to be set", HostRootEnv)
	}

	hostRootAbs, err := filepath.Abs(hostRoot)

	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", HostRootEnv, err)
	}

	resolved := make([]string, 0, len(h))
	seen := map[string]struct{}{}

	for _, hostPath := range h {
		pathValue := string(hostPath)

		if !filepath.IsAbs(pathValue) {
			return nil, fmt.Errorf("--host-path must be an absolute path")
		}

		hostPathAbs, err := filepath.Abs(pathValue)

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

		mapped := workRoot

		if rel != "." {
			mapped = filepath.Join(workRoot, rel)
		}

		if _, ok := seen[mapped]; ok {
			continue
		}

		resolved = append(resolved, mapped)
		seen[mapped] = struct{}{}
	}

	return resolved, nil
}
