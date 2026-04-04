package engine_test

import (
	"path/filepath"
	"testing"

	"github.com/oullin/go-fmt/semantic/engine"
)

func TestResolveRunPathsRootMapsToWorkRoot(t *testing.T) {
	workRoot := filepath.Join(string(filepath.Separator), "work")
	hostRoot := filepath.Join(string(filepath.Separator), "host", "project")

	t.Setenv(engine.HostRootEnv, hostRoot)

	paths, err := engine.HostPath(hostRoot).Resolve(workRoot, nil)

	if err != nil {
		t.Fatalf("resolve host path: %v", err)
	}

	if len(paths) != 1 || paths[0] != workRoot {
		t.Fatalf("unexpected paths: %#v", paths)
	}
}

func TestResolveRunPathsNestedDirectoryMapsToWorkRoot(t *testing.T) {
	workRoot := filepath.Join(string(filepath.Separator), "work")
	hostRoot := filepath.Join(string(filepath.Separator), "host", "project")
	hostPath := filepath.Join(hostRoot, "pkg", "api")

	t.Setenv(engine.HostRootEnv, hostRoot)

	paths, err := engine.HostPath(hostPath).Resolve(workRoot, nil)

	if err != nil {
		t.Fatalf("resolve host path: %v", err)
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

	t.Setenv(engine.HostRootEnv, hostRoot)

	paths, err := engine.HostPath(hostPath).Resolve(workRoot, nil)

	if err != nil {
		t.Fatalf("resolve host path: %v", err)
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

	t.Setenv(engine.HostRootEnv, hostRoot)

	_, err := engine.HostPath(hostPath).Resolve(workRoot, nil)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveRunPathsRejectsMissingHostRoot(t *testing.T) {
	workRoot := filepath.Join(string(filepath.Separator), "work")
	hostPath := filepath.Join(string(filepath.Separator), "host", "project")

	t.Setenv(engine.HostRootEnv, "")

	_, err := engine.HostPath(hostPath).Resolve(workRoot, nil)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveRunPathsRejectsPositionalPaths(t *testing.T) {
	workRoot := filepath.Join(string(filepath.Separator), "work")
	hostRoot := filepath.Join(string(filepath.Separator), "host", "project")

	t.Setenv(engine.HostRootEnv, hostRoot)

	_, err := engine.HostPath(hostRoot).Resolve(workRoot, []string{"."})

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveRunPathsSupportsMultipleHostPaths(t *testing.T) {
	workRoot := filepath.Join(string(filepath.Separator), "work")
	hostRoot := filepath.Join(string(filepath.Separator), "host", "project")
	first := filepath.Join(hostRoot, "pkg", "api")
	second := filepath.Join(hostRoot, "internal", "app")

	t.Setenv(engine.HostRootEnv, hostRoot)

	paths, err := engine.HostPaths{engine.HostPath(first), engine.HostPath(second)}.Resolve(workRoot, nil)

	if err != nil {
		t.Fatalf("resolve host paths: %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %#v", paths)
	}

	if paths[0] != filepath.Join(workRoot, "pkg", "api") {
		t.Fatalf("unexpected first path: %q", paths[0])
	}

	if paths[1] != filepath.Join(workRoot, "internal", "app") {
		t.Fatalf("unexpected second path: %q", paths[1])
	}
}
