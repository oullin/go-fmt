package formatter_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/go-fmt/packages/driver/testutil"
	"github.com/oullin/go-fmt/packages/formatter"
	"github.com/oullin/go-fmt/packages/formatter/config"
)

func TestCheckReportsStyleChanges(t *testing.T) {
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

	report, err := formatter.Check([]string{root}, config.Default())

	if err != nil {
		t.Fatalf("check: %v", err)
	}

	if report.Result != "fail" {
		t.Fatalf("expected fail result, got %q", report.Result)
	}
}

func TestFormatWritesChanges(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.go")
	testutil.WriteGoFile(t, path, `package sample

func run() {
	defer println("done")
	return
}
`)

	report, err := formatter.Format([]string{root}, config.Default())

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
		t.Fatalf("expected formatted file, got:\n%s", content)
	}
}

func TestFormatRepairsGoEmbedDirectivePlacement(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.go")
	testutil.WriteGoFile(t, path, `package sample

import "embed"

//go:embed foo.txt

type runtime struct{}

var rootTemplateFS embed.FS
`)

	report, err := formatter.Format([]string{root}, config.Default())

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

	expected := `package sample

import "embed"

//go:embed foo.txt
var rootTemplateFS embed.FS

type runtime struct{}
`

	if string(content) != expected {
		t.Fatalf("expected repaired go:embed placement, got:\n%s", content)
	}
}
