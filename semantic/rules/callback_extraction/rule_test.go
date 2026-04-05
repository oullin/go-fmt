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

func TestApplySkipsDirectLabeledStatements(t *testing.T) {
	path := writeTempGoFile(t, `package sample

type handler struct {
	Redirect func(string)
}

func consume(handler) {}

func run() {
	goto L
L:
	consume(handler{
		Redirect: func(url string) {
			println(url)
		},
	})
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 0 {
		t.Fatalf("expected 0 violations for direct labeled statement, got %d", len(violations))
	}

	output := string(formatted)

	if strings.Contains(output, "redirectFn := func(url string)") {
		t.Fatalf("expected labeled statement to keep inline callback, got:\n%s", formatted)
	}

	if !strings.Contains(output, "L:\n\tconsume(handler{") {
		t.Fatalf("expected callback to stay in labeled statement, got:\n%s", formatted)
	}

	if !strings.Contains(output, "Redirect: func(url string)") {
		t.Fatalf("expected inline callback to remain under label, got:\n%s", formatted)
	}
}

func TestApplyExtractsInsideNestedLabeledBodies(t *testing.T) {
	path := writeTempGoFile(t, `package sample

type handler struct {
	Redirect func(string)
}

func run() {
L:
	for {
		h := handler{
			Redirect: func(url string) {
				println(url)
			},
		}

		_ = h
		break L
	}
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation in nested labeled body, got %d", len(violations))
	}

	output := string(formatted)

	if !strings.Contains(output, "L:\n\tfor {\n\t\tredirectFn := func(url string)") {
		t.Fatalf("expected extraction inside labeled loop body, got:\n%s", formatted)
	}

	if !strings.Contains(output, "Redirect: redirectFn") {
		t.Fatalf("expected extracted callback reference in labeled loop body, got:\n%s", formatted)
	}
}

func TestLowerFirst(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "ascii title case",
			input: "Redirect",
			want:  "redirect",
		},
		{
			name:  "unicode title case",
			input: "Éclair",
			want:  "éclair",
		},
		{
			name:  "whitespace only",
			input: " \t\n ",
			want:  "callback",
		},
		{
			name:  "already lowercase",
			input: "redirect",
			want:  "redirect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lowerFirst(tt.input); got != tt.want {
				t.Fatalf("lowerFirst(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
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
