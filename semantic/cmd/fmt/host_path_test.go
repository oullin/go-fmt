package main

import (
	"path/filepath"
	"testing"
)

func TestResolveRunPathsRootMapsToWorkRoot(t *testing.T) {
	workRoot := filepath.Join(string(filepath.Separator), "work")
	hostRoot := filepath.Join(string(filepath.Separator), "host", "project")
	t.Setenv(hostRootEnv, hostRoot)

	paths, err := resolveRunPaths(workRoot, hostRoot, nil)

	if err != nil {
		t.Fatalf("resolveRunPaths: %v", err)
	}

	if len(paths) != 1 || paths[0] != workRoot {
		t.Fatalf("unexpected paths: %#v", paths)
	}
}

func TestResolveRunPathsNestedDirectoryMapsToWorkRoot(t *testing.T) {
	workRoot := filepath.Join(string(filepath.Separator), "work")
	hostRoot := filepath.Join(string(filepath.Separator), "host", "project")
	hostPath := filepath.Join(hostRoot, "pkg", "api")
	t.Setenv(hostRootEnv, hostRoot)

	paths, err := resolveRunPaths(workRoot, hostPath, nil)

	if err != nil {
		t.Fatalf("resolveRunPaths: %v", err)
	}

	want := filepath.Join(workRoot, "pkg", "api")

	if len(paths) != 1 || paths[0] != want {
		t.Fatalf("unexpected paths: got %#v want %q", paths, want)
	}
}

func TestResolveRunPathsSingleFileMapsToWorkRoot(t *testing.T) {
	workRoot := filepath.Join(string(filepath.Separator), "work")
	hostRoot := filepath.Join(string(filepath.Separator), "host", "project")
	hostPath := filepath.Join(hostRoot, "pkg", "api", "sample.go")
	t.Setenv(hostRootEnv, hostRoot)

	paths, err := resolveRunPaths(workRoot, hostPath, nil)

	if err != nil {
		t.Fatalf("resolveRunPaths: %v", err)
	}

	want := filepath.Join(workRoot, "pkg", "api", "sample.go")

	if len(paths) != 1 || paths[0] != want {
		t.Fatalf("unexpected paths: got %#v want %q", paths, want)
	}
}

func TestResolveRunPathsRejectsOutsideRoot(t *testing.T) {
	workRoot := filepath.Join(string(filepath.Separator), "work")
	hostRoot := filepath.Join(string(filepath.Separator), "host", "project")
	hostPath := filepath.Join(string(filepath.Separator), "host", "other")
	t.Setenv(hostRootEnv, hostRoot)

	_, err := resolveRunPaths(workRoot, hostPath, nil)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveRunPathsRejectsMissingHostRoot(t *testing.T) {
	workRoot := filepath.Join(string(filepath.Separator), "work")
	hostPath := filepath.Join(string(filepath.Separator), "host", "project")
	t.Setenv(hostRootEnv, "")

	_, err := resolveRunPaths(workRoot, hostPath, nil)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveRunPathsRejectsPositionalPaths(t *testing.T) {
	workRoot := filepath.Join(string(filepath.Separator), "work")
	hostRoot := filepath.Join(string(filepath.Separator), "host", "project")
	t.Setenv(hostRootEnv, hostRoot)

	_, err := resolveRunPaths(workRoot, hostRoot, []string{"."})

	if err == nil {
		t.Fatal("expected error")
	}
}
