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

func TestApplySkipsShadowedDefaultImport(t *testing.T) {
	source := `package sample

import "strings"

func run(strings string) bool {
	return strings != ""
}
`
	path := writeTempGoFile(t, source)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}

	if string(formatted) != source {
		t.Fatalf("expected source to be unchanged, got:\n%s", formatted)
	}
}

func TestApplySkipsShadowedMissingImport(t *testing.T) {
	source := `package sample

func run(strings string) bool {
	return strings != ""
}
`
	path := writeTempGoFile(t, source)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}

	if strings.Contains(string(formatted), "import \"strings\"") {
		t.Fatalf("expected no strings import, got:\n%s", formatted)
	}

	if string(formatted) != source {
		t.Fatalf("expected source to be unchanged, got:\n%s", formatted)
	}
}

func TestApplyRewritesSafeSitesAndSkipsShadowedSites(t *testing.T) {
	path := writeTempGoFile(t, `package sample

func run(value string) bool {
	if value != "" {
		return true
	}

	return func(strings string) bool {
		return strings != ""
	}(value)
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

	if !strings.Contains(string(formatted), "if strings.TrimSpace(value) != \"\"") {
		t.Fatalf("expected safe site rewrite, got:\n%s", formatted)
	}

	if !strings.Contains(string(formatted), "return strings != \"\"") {
		t.Fatalf("expected shadowed site to remain unchanged, got:\n%s", formatted)
	}
}

func TestApplyReusesAliasWhenStringsIsShadowedLocally(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import stdstrings "strings"

func run(value string) bool {
	return func(strings string) bool {
		return strings != "" || value != ""
	}(value)
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "return stdstrings.TrimSpace(strings) != \"\" || stdstrings.TrimSpace(value) != \"\"") {
		t.Fatalf("expected alias-based rewrite, got:\n%s", formatted)
	}
}

func TestApplyPreservesDotImport(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import . "strings"

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

	if !strings.Contains(string(formatted), "import . \"strings\"") {
		t.Fatalf("expected dot import to be preserved, got:\n%s", formatted)
	}

	if !strings.Contains(string(formatted), "return TrimSpace(value) != \"\"") {
		t.Fatalf("expected dot import rewrite, got:\n%s", formatted)
	}
}

func TestApplySkipsBlankImport(t *testing.T) {
	source := `package sample

import _ "strings"

func run(value string) bool {
	return value != ""
}
`
	path := writeTempGoFile(t, source)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}

	if string(formatted) != source {
		t.Fatalf("expected blank import source to be unchanged, got:\n%s", formatted)
	}
}

func TestApplyDoesNotWrapExistingTrimSpaceCalls(t *testing.T) {
	testCases := []struct {
		name   string
		source string
	}{
		{
			name: "default import",
			source: `package sample

import "strings"

func run(value string) bool {
	return strings.TrimSpace(value) != ""
}
`,
		},
		{
			name: "alias import",
			source: `package sample

import stdstrings "strings"

func run(value string) bool {
	return stdstrings.TrimSpace(value) != ""
}
`,
		},
		{
			name: "dot import",
			source: `package sample

import . "strings"

func run(value string) bool {
	return TrimSpace(value) != ""
}
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTempGoFile(t, tc.source)

			violations, formatted, err := New().Apply(path, mustReadFile(t, path))

			if err != nil {
				t.Fatalf("apply: %v", err)
			}

			if len(violations) != 0 {
				t.Fatalf("expected no violations, got %d", len(violations))
			}

			if string(formatted) != tc.source {
				t.Fatalf("expected source to be unchanged, got:\n%s", formatted)
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
