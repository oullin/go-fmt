package spacing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyFindsMissingBlankLineAfterIf(t *testing.T) {
	path := writeTempGoFile(t, `package sample

func run() {
	if true {
		println("ok")
	}
	println("next")
}
`)

	violations, _, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(violations[0].Message, "after if statement") {
		t.Fatalf("unexpected message %q", violations[0].Message)
	}
}

func TestApplyFormatsDeferSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

func run() {
	defer println("done")
	return
}
`)

	_, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if !strings.Contains(string(formatted), "defer println(\"done\")\n\n\treturn") {
		t.Fatalf("expected blank line after defer, got:\n%s", formatted)
	}
}

func TestApplyChecksCaseBodies(t *testing.T) {
	path := writeTempGoFile(t, `package sample

func run(v int) {
	switch v {
	case 1:
		if true {
			println("ok")
		}
		println("next")
	}
}
`)

	violations, _, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
}

func TestApplyFlagsTypeDefinitionsAfterFunctions(t *testing.T) {
	path := writeTempGoFile(t, `package sample

func run() {}

type later struct{}
`)

	violations, _, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(violations[0].Message, "type definitions must appear") {
		t.Fatalf("unexpected message %q", violations[0].Message)
	}
}

func TestApplyFormatsTypeDefinitionsAtBeginningOfFile(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "fmt"

func run() {
	fmt.Println("ok")
}

// later docs
type later struct{}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	want := `package sample

import "fmt"

// later docs
type later struct{}

func run() {
	fmt.Println("ok")
}
`

	if string(formatted) != want {
		t.Fatalf("unexpected formatted output:\n%s", formatted)
	}
}

func TestApplyFindsVarSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

func run() {
	println("start")
	var total int
	total++
}
`)

	violations, _, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}
}

func TestApplyFindsContinueSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

func run() {
	for i := 0; i < 10; i++ {
		println(i)
		if i % 2 == 0 {
			println("even")
			continue
		}
		println("odd")
	}
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	found := false

	for _, v := range violations {
		if strings.Contains(v.Message, "before continue") {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("expected violation for 'before continue', got %d violations: %v", len(violations), violations)
	}

	if !strings.Contains(string(formatted), "\n\n\t\t\tcontinue") {
		t.Errorf("expected blank line before continue in formatted output:\n%s", formatted)
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
