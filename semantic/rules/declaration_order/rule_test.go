package declaration_order

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyMovesVarsBeforeTypes(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "fmt"

type config struct{}

var defaultName = "ok"

func run() {
	fmt.Println(defaultName)
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "var defaultName = \"ok\"\n\ntype config struct{}") {
		t.Fatalf("expected vars before types, got:\n%s", formatted)
	}
}

func TestApplyPreservesGoEmbedAdjacency(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "embed"

//go:embed foo.txt

type runtime struct{}

var rootTemplateFS embed.FS
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "//go:embed foo.txt\nvar rootTemplateFS embed.FS") {
		t.Fatalf("expected embed directive to stay attached to var, got:\n%s", formatted)
	}
}

func writeTempGoFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	return path
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	content, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	return content
}
