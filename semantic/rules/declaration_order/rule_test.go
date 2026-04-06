package declaration_order

import (
	"bytes"
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

func TestApplyMovesConstsBeforeVarsAndTypes(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "fmt"

var defaultName = "ok"

type config struct{}

const version = "v1"

func run() {
	fmt.Println(defaultName, version)
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "const version = \"v1\"\n\nvar defaultName = \"ok\"\n\ntype config struct{}") {
		t.Fatalf("expected consts before vars and types, got:\n%s", formatted)
	}
}

func TestApplyMovesConstsAboveFunctionsWithoutVars(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "fmt"

func run() {
	fmt.Println(version)
}

const version = "v1"
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "import \"fmt\"\n\nconst version = \"v1\"\n\nfunc run()") {
		t.Fatalf("expected consts before functions even without vars, got:\n%s", formatted)
	}
}

func TestApplyNormalizesTypeConstVarOrder(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "fmt"

type config struct{}

const version = "v1"

var defaultName = "ok"

func run() {
	fmt.Println(defaultName, version)
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "const version = \"v1\"\n\nvar defaultName = \"ok\"\n\ntype config struct{}") {
		t.Fatalf("expected const-var-type order, got:\n%s", formatted)
	}
}

func TestApplyLeavesOrderedConstVarTypeDeclarations(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "fmt"

const version = "v1"

var defaultName = "ok"

type config struct{}

func run() {
	fmt.Println(defaultName, version)
}
`)

	input := mustReadFile(t, path)

	violations, formatted, err := New().Apply(path, input)

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}

	if !bytes.Equal(formatted, input) {
		t.Fatalf("expected already ordered file to remain unchanged, got:\n%s", formatted)
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

	expected := `package sample

import "embed"

//go:embed foo.txt
var rootTemplateFS embed.FS

type runtime struct{}
`

	if !bytes.Equal(formatted, []byte(expected)) {
		t.Fatalf("expected embed directive repair output, got:\n%s", formatted)
	}
}

func TestApplyLeavesOrderedGoEmbedDirectiveUnchanged(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "embed"

//go:embed foo.txt
var rootTemplateFS embed.FS

type runtime struct{}
`)

	input := mustReadFile(t, path)

	violations, formatted, err := New().Apply(path, input)

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}

	if !bytes.Equal(formatted, input) {
		t.Fatalf("expected already attached go:embed declaration to remain unchanged, got:\n%s", formatted)
	}
}

func TestApplyPreservesGoEmbedAdjacencyWithTab(t *testing.T) {
	path := writeTempGoFile(t, "package sample\n\nimport \"embed\"\n\n//go:embed\tfoo.txt\n\ntype runtime struct{}\n\nvar rootTemplateFS embed.FS\n")

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	expected := "package sample\n\nimport \"embed\"\n\n//go:embed\tfoo.txt\nvar rootTemplateFS embed.FS\n\ntype runtime struct{}\n"

	if !bytes.Equal(formatted, []byte(expected)) {
		t.Fatalf("expected embed directive repair output, got:\n%s", formatted)
	}
}

func TestApplyLeavesOrderedGoEmbedDirectiveWithTabUnchanged(t *testing.T) {
	path := writeTempGoFile(t, "package sample\n\nimport \"embed\"\n\n//go:embed\tfoo.txt\nvar rootTemplateFS embed.FS\n\ntype runtime struct{}\n")

	input := mustReadFile(t, path)

	violations, formatted, err := New().Apply(path, input)

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}

	if !bytes.Equal(formatted, input) {
		t.Fatalf("expected already attached go:embed declaration to remain unchanged, got:\n%s", formatted)
	}
}

func TestCollapseEmbedSpacingRecognizesVarForms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{
			name: "single line var",
			src: `package sample

//go:embed foo.txt

var rootTemplateFS embed.FS
`,
			expected: `package sample

//go:embed foo.txt
var rootTemplateFS embed.FS
`,
		},
		{
			name: "grouped var without space",
			src: `package sample

//go:embed foo.txt

var(
	rootTemplateFS embed.FS
)
`,
			expected: `package sample

//go:embed foo.txt
var(
	rootTemplateFS embed.FS
)
`,
		},
		{
			name: "grouped var with multiple spaces",
			src: `package sample

//go:embed foo.txt

var    (
	rootTemplateFS embed.FS
)
`,
			expected: `package sample

//go:embed foo.txt
var    (
	rootTemplateFS embed.FS
)
`,
		},
		{
			name: "non var declaration",
			src: `package sample

//go:embed foo.txt

type runtime struct{}
`,
			expected: `package sample

//go:embed foo.txt

type runtime struct{}
`,
		},
		{
			name:     "tab separated directive",
			src:      "package sample\n\n//go:embed\tfoo.txt\n\nvar rootTemplateFS embed.FS\n",
			expected: "package sample\n\n//go:embed\tfoo.txt\nvar rootTemplateFS embed.FS\n",
		},
		{
			name: "bare directive",
			src: `package sample

//go:embed

var rootTemplateFS embed.FS
`,
			expected: `package sample

//go:embed

var rootTemplateFS embed.FS
`,
		},
		{
			name: "embedded prefix",
			src: `package sample

//go:embedded foo.txt

var rootTemplateFS embed.FS
`,
			expected: `package sample

//go:embedded foo.txt

var rootTemplateFS embed.FS
`,
		},
		{
			name: "identifier starting with var",
			src: `package sample

//go:embed foo.txt

variant := 1
`,
			expected: `package sample

//go:embed foo.txt

variant := 1
`,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			formatted := collapseEmbedSpacing([]byte(tt.src))

			if !bytes.Equal(formatted, []byte(tt.expected)) {
				t.Fatalf("unexpected output:\n%s", formatted)
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
