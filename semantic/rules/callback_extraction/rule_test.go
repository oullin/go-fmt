package callback_extraction

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyExtractsCompositeLiteralCallbacks(t *testing.T) {
	path := writeTempGoFile(t, `package sample

type handler struct {
	Redirect func(string)
}

func run() {
	h := handler{
		Redirect: func(url string) {
			println(url)
		},
	}

	_ = h
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "redirectFn := func(url string)") {
		t.Fatalf("expected extracted callback variable, got:\n%s", formatted)
	}

	if !strings.Contains(string(formatted), "Redirect: redirectFn") {
		t.Fatalf("expected field to reuse extracted callback, got:\n%s", formatted)
	}
}

func TestApplyAvoidsNameCollisions(t *testing.T) {
	path := writeTempGoFile(t, `package sample

type handler struct {
	Redirect func(string)
}

func run() {
	redirectFn := "taken"
	h := handler{
		Redirect: func(url string) {
			println(redirectFn, url)
		},
	}

	_ = h
}
`)

	_, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if !strings.Contains(string(formatted), "redirectFn1 := func(url string)") {
		t.Fatalf("expected deterministic suffix for collision, got:\n%s", formatted)
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
