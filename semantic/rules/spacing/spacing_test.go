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

func TestApplyFormatsBlankLineBeforeSortCall(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "sort"

func run(values []string) {
	println("start")
	sort.Strings(values)
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(violations[0].Message, "before sort call") {
		t.Fatalf("unexpected message %q", violations[0].Message)
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\tsort.Strings(values)") {
		t.Fatalf("expected blank line before sort call, got:\n%s", formatted)
	}
}

func TestApplyFormatsBlankLineAfterSortCall(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "sort"

func run(values []string) {
	sort.Strings(values)
	println("done")
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(violations[0].Message, "after sort call") {
		t.Fatalf("unexpected message %q", violations[0].Message)
	}

	if !strings.Contains(string(formatted), "sort.Strings(values)\n\n\tprintln(\"done\")") {
		t.Fatalf("expected blank line after sort call, got:\n%s", formatted)
	}
}

func TestApplyFormatsBlankLinesAroundSortCall(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "sort"

func run(values []string) {
	println("start")
	sort.Strings(values)
	println("done")
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\tsort.Strings(values)\n\n\tprintln(\"done\")") {
		t.Fatalf("expected blank lines around sort call, got:\n%s", formatted)
	}
}

func TestApplyFormatsAliasedSortCallSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import stdsort "sort"

func run(values []string) {
	println("start")
	stdsort.Strings(values)
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\tstdsort.Strings(values)") {
		t.Fatalf("expected blank line before aliased sort call, got:\n%s", formatted)
	}
}

func TestApplyFormatsSlicesSortCallSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "slices"

func run(values []int) {
	println("start")
	slices.Sort(values)
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\tslices.Sort(values)") {
		t.Fatalf("expected blank line before slices sort call, got:\n%s", formatted)
	}
}

func TestApplyFormatsAliasedSlicesSortFuncSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import stdslices "slices"

func run(values []int) {
	println("start")
	stdslices.SortFunc(values, func(a, b int) int {
		return a - b
	})
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\tstdslices.SortFunc(values, func(a, b int) int {") {
		t.Fatalf("expected blank line before aliased slices sort call, got:\n%s", formatted)
	}
}

func TestApplyFormatsBlankLineBeforeRandCall(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "math/rand"

func run() {
	println("start")
	rand.Int()
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(violations[0].Message, "before rand call") {
		t.Fatalf("unexpected message %q", violations[0].Message)
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\trand.Int()") {
		t.Fatalf("expected blank line before rand call, got:\n%s", formatted)
	}
}

func TestApplyFormatsBlankLineAfterRandCall(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "math/rand"

func run() {
	rand.Int()
	println("done")
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(violations[0].Message, "after rand call") {
		t.Fatalf("unexpected message %q", violations[0].Message)
	}

	if !strings.Contains(string(formatted), "rand.Int()\n\n\tprintln(\"done\")") {
		t.Fatalf("expected blank line after rand call, got:\n%s", formatted)
	}
}

func TestApplyFormatsBlankLinesAroundRandCall(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "math/rand"

func run() {
	println("start")
	rand.Int()
	println("done")
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\trand.Int()\n\n\tprintln(\"done\")") {
		t.Fatalf("expected blank lines around rand call, got:\n%s", formatted)
	}
}

func TestApplyFormatsAliasedRandCallSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import stdrand "math/rand"

func run() {
	println("start")
	stdrand.Int()
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\tstdrand.Int()") {
		t.Fatalf("expected blank line before aliased rand call, got:\n%s", formatted)
	}
}

func TestApplyFormatsRandV2CallSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "math/rand/v2"

func run() {
	println("start")
	rand.Int()
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\trand.Int()") {
		t.Fatalf("expected blank line before rand/v2 call, got:\n%s", formatted)
	}
}

func TestApplyFormatsAliasedRandV2CallSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import random "math/rand/v2"

func run() {
	println("start")
	random.Int()
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\trandom.Int()") {
		t.Fatalf("expected blank line before aliased rand/v2 call, got:\n%s", formatted)
	}
}

func TestApplyIgnoresNonStandaloneSortUsage(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "sort"

func run(values []string) {
	println(sort.StringsAreSorted(values))
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}

	if string(formatted) != string(mustReadFile(t, path)) {
		t.Fatalf("expected unchanged output, got:\n%s", formatted)
	}
}

func TestApplyIgnoresNonStandaloneRandUsage(t *testing.T) {
	path := writeTempGoFile(t, `package sample

import "math/rand"

func run() {
	println(rand.Int())
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}

	if string(formatted) != string(mustReadFile(t, path)) {
		t.Fatalf("expected unchanged output, got:\n%s", formatted)
	}
}

func TestApplyFormatsBlankLinesAroundGenericSelectorCalls(t *testing.T) {
	path := writeTempGoFile(t, `package sample

type sorter struct{}

func (sorter) Strings([]string) {}

func run(values []string) {
	println("start")
	sort := sorter{}
	sort.Strings(values)
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\tsort := sorter{}\n\n\tsort.Strings(values)") {
		t.Fatalf("expected generic selector call spacing, got:\n%s", formatted)
	}
}

func TestApplyFormatsBlankLinesAroundGenericMethodCalls(t *testing.T) {
	path := writeTempGoFile(t, `package sample

type randomizer struct{}

func (randomizer) Int() int { return 1 }

func run() {
	println("start")
	rand := randomizer{}
	rand.Int()
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\trand := randomizer{}\n\n\trand.Int()") {
		t.Fatalf("expected generic method call spacing, got:\n%s", formatted)
	}
}

func TestApplyFormatsBlankLineBeforeGenericCallAfterAssignment(t *testing.T) {
	path := writeTempGoFile(t, `package sample

type sliceOps struct{}

func (sliceOps) Sort([]int) {}

func run(values []int) {
	println("start")
	slices := sliceOps{}
	slices.Sort(values)
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\tslices := sliceOps{}\n\n\tslices.Sort(values)") {
		t.Fatalf("expected generic call spacing, got:\n%s", formatted)
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
