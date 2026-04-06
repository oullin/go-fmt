package engine_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/go-fmt/packages/semantic/config"
	"github.com/oullin/go-fmt/packages/semantic/engine"
	"github.com/oullin/go-fmt/packages/semantic/formatter"
	"github.com/oullin/go-fmt/packages/semantic/rules"
	"github.com/oullin/go-fmt/packages/semantic/rules/spacing"
	"github.com/oullin/go-fmt/packages/testutil"
)

func defaultRules() []rules.Rule {
	return []rules.Rule{spacing.New()}
}

func defaultFormatters() []formatter.Formatter {
	return []formatter.Formatter{formatter.NewGofmt(), formatter.NewGoimports()}
}

func TestCollectGoFilesSkipsHiddenVendorAndGenerated(t *testing.T) {
	root := t.TempDir()
	testutil.WriteGoFile(t, filepath.Join(root, "root.go"), "package sample\n")
	testutil.WriteGoFile(t, filepath.Join(root, "pkg", "nested.go"), "package sample\n")
	testutil.WriteGoFile(t, filepath.Join(root, "vendor", "skip.go"), "package sample\n")
	testutil.WriteGoFile(t, filepath.Join(root, ".hidden", "skip.go"), "package sample\n")
	testutil.WriteGoFile(t, filepath.Join(root, "generated.gen.go"), "package sample\n")

	files, err := engine.CollectGoFiles([]string{root}, config.Default())

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
	testutil.WriteGoFile(t, path, `package sample

func run() {
	if true {
		println("ok")
	}
	println("next")
}
`)

	report, err := engine.New(config.Default(), defaultRules(), defaultFormatters()).Check([]string{root})

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
	testutil.WriteGoFile(t, path, `package sample

func run() {
	defer println("done")
	return
}
`)

	report, err := engine.New(config.Default(), defaultRules(), defaultFormatters()).Format([]string{root})

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

func TestFormatSkipsSingleLineFuncLiteralSpacingViolations(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.go")
	testutil.WriteGoFile(t, path, `package sample

type config struct {
	SecureCookie bool
}

func run() config {
	return config{
		SecureCookie: func() bool { value := true; return value }(),
	}
}
`)

	report, err := engine.New(config.Default(), defaultRules(), defaultFormatters()).Format([]string{root})

	if err != nil {
		t.Fatalf("format: %v", err)
	}

	if report.Result != "pass" {
		t.Fatalf("expected pass result, got %q", report.Result)
	}

	if report.Changed != 0 {
		t.Fatalf("expected 0 changed files, got %d", report.Changed)
	}

	if report.ViolationCount() != 0 {
		t.Fatalf("expected 0 violations, got %d", report.ViolationCount())
	}
}
