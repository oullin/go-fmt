package trimspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyWrapsStringComparisonsAndAddsImport(t *testing.T) {
	path := writeTempGoFile(t, `package sample

func run(value string) bool {
	return value != ""
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "import \"strings\"") {
		t.Fatalf("expected strings import, got:\n%s", formatted)
	}

	if !strings.Contains(string(formatted), "return strings.TrimSpace(value) != \"\"") {
		t.Fatalf("expected TrimSpace rewrite, got:\n%s", formatted)
	}
}

func TestApplyReusesExistingAlias(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import stdstrings "strings"

func run(value string) bool {
	if current := value; "" == current {
		return true
	}

	return false
}
`)

	_, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if !strings.Contains(string(formatted), "\"\" == stdstrings.TrimSpace(current)") {
		t.Fatalf("expected alias reuse, got:\n%s", formatted)
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
