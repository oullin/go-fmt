package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/go-fmt/internal/config"
)

func TestCollectGoFilesSkipsHiddenVendorAndGenerated(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "root.go"), "package sample\n")
	mustWrite(t, filepath.Join(root, "pkg", "nested.go"), "package sample\n")
	mustWrite(t, filepath.Join(root, "vendor", "skip.go"), "package sample\n")
	mustWrite(t, filepath.Join(root, ".hidden", "skip.go"), "package sample\n")
	mustWrite(t, filepath.Join(root, "generated.gen.go"), "package sample\n")

	files, err := CollectGoFiles([]string{root}, config.Default())

	if err != nil {
		t.Fatalf("collect files: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %#v", len(files), files)
	}
}

func TestCheckReportsStyleChangesWithoutWriting(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.go")
	mustWrite(t, path, `package sample

func run() {
	if true {
		println("ok")
	}
	println("next")
}
`)

	report, err := New(config.Default()).Check([]string{root})

	if err != nil {
		t.Fatalf("check: %v", err)
	}

	if report.Result != "fail" {
		t.Fatalf("expected fail result, got %q", report.Result)
	}

	if report.Changed != 1 {
		t.Fatalf("expected 1 changed file, got %d", report.Changed)
	}

	content, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if strings.Contains(string(content), "\n\n\tprintln(\"next\")") {
		t.Fatalf("check should not write changes")
	}
}

func TestFormatWritesSpacingChanges(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.go")
	mustWrite(t, path, `package sample

func run() {
	defer println("done")
	return
}
`)

	report, err := New(config.Default()).Format([]string{root})

	if err != nil {
		t.Fatalf("format: %v", err)
	}

	if report.Result != "fixed" {
		t.Fatalf("expected fixed result, got %q", report.Result)
	}

	content, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if !strings.Contains(string(content), "defer println(\"done\")\n\n\treturn") {
		t.Fatalf("expected file to be rewritten, got:\n%s", content)
	}
}

func mustWrite(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
