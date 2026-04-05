package spacing

import (
	"strings"
	"testing"
)

func TestApplyFindsConstSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

func run() {
	println("start")
	const answer = 42
	println(answer)
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\tconst answer = 42\n\n\tprintln(answer)") {
		t.Fatalf("expected const spacing, got:\n%s", formatted)
	}
}

func TestApplyFindsFunctionAssignmentSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

func run() {
	println("start")
	handler := func() {
		println("ok")
	}
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

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\thandler := func()") {
		t.Fatalf("expected blank line before function assignment, got:\n%s", formatted)
	}

	if !strings.Contains(string(formatted), "}\n\n\tprintln(\"done\")") {
		t.Fatalf("expected blank line after function assignment, got:\n%s", formatted)
	}
}

func TestApplyFindsRouteGroupAndMiddlewareSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

func run(route interface{ Group(string, func()) }, next interface{ ServeHTTP() }) {
	println("start")
	route.Group("/ok", func() {})
	next.ServeHTTP()
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "println(\"start\")\n\n\troute.Group") {
		t.Fatalf("expected blank line before route group call, got:\n%s", formatted)
	}

	if !strings.Contains(string(formatted), "func() {})\n\n\tnext.ServeHTTP()") {
		t.Fatalf("expected blank line before next middleware call, got:\n%s", formatted)
	}
}

func TestApplyFindsMutexReadLockAndUnlockSpacing(t *testing.T) {
	path := writeTempGoFile(t, `package sample

type locker struct {
	mu mutex
}

type mutex struct{}

func (mutex) RLock() {}
func (mutex) RUnlock() {}

func run(l locker) {
	l.mu.RLock()
	println("locked")
	l.mu.RUnlock()
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "l.mu.RLock()\n\n\tprintln(\"locked\")") {
		t.Fatalf("expected blank line after RLock, got:\n%s", formatted)
	}

	if !strings.Contains(string(formatted), "println(\"locked\")\n\n\tl.mu.RUnlock()") {
		t.Fatalf("expected blank line before RUnlock, got:\n%s", formatted)
	}
}

func TestApplyFindsMutexReadLockAndUnlockSpacingForIdentifierReceiver(t *testing.T) {
	path := writeTempGoFile(t, `package sample

type mutex struct{}

func (mutex) RLock() {}
func (mutex) RUnlock() {}

func run() {
	var mu mutex
	mu.RLock()
	println("locked")
	mu.RUnlock()
}
`)

	violations, formatted, err := New().Apply(path, mustReadFile(t, path))

	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(violations) != 3 {
		t.Fatalf("expected 3 violations, got %d", len(violations))
	}

	if !strings.Contains(string(formatted), "var mu mutex\n\n\tmu.RLock()") {
		t.Fatalf("expected blank line before RLock, got:\n%s", formatted)
	}

	if !strings.Contains(string(formatted), "mu.RLock()\n\n\tprintln(\"locked\")") {
		t.Fatalf("expected blank line after RLock, got:\n%s", formatted)
	}

	if !strings.Contains(string(formatted), "println(\"locked\")\n\n\tmu.RUnlock()") {
		t.Fatalf("expected blank line before RUnlock, got:\n%s", formatted)
	}
}
