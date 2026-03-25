package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/go-fmt/semantic/engine"
)

func TestRunCheckFailsOnStyleChanges(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "sample.go"), `package sample

func run() {
	if true {
		println("ok")
	}
	println("next")
}
`)

	exitCode, stdout, stderr := runCLI(t, dir, "check", dir)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Result: fail") {
		t.Fatalf("unexpected stdout:\n%s", stdout)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
}

func TestRunFormatWritesChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	mustWrite(t, path, `package sample

func run() {
	defer println("done")
	return
}
`)

	exitCode, stdout, stderr := runCLI(t, dir, "format", dir)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Result: fixed") {
		t.Fatalf("unexpected stdout:\n%s", stdout)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}

	content, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if !strings.Contains(string(content), "defer println(\"done\")\n\n\treturn") {
		t.Fatalf("expected formatted file, got:\n%s", content)
	}
}

func TestRunAgentOutput(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "sample.go"), `package sample

func run() {
	if true {
		println("ok")
	}
	println("next")
}
`)

	exitCode, stdout, stderr := runCLI(t, dir, "check", "--format", "agent", dir)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stdout, "\"violations\"") || !strings.Contains(stdout, "\"changed\"") {
		t.Fatalf("unexpected agent output:\n%s", stdout)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
}

func TestRunCheckWithHostPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	mustWrite(t, path, `package sample

func run() {
	if true {
		println("ok")
	}
	println("next")
}
`)
	t.Setenv(engine.HostRootEnv, dir)

	exitCode, stdout, stderr := runCLI(t, dir, "check", "--host-path", path)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Result: fail") {
		t.Fatalf("unexpected stdout:\n%s", stdout)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
}

func TestRunFormatWithHostPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	mustWrite(t, path, `package sample

func run() {
	defer println("done")
	return
}
`)
	t.Setenv(engine.HostRootEnv, dir)

	exitCode, stdout, stderr := runCLI(t, dir, "format", "--host-path", path)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Result: fixed") {
		t.Fatalf("unexpected stdout:\n%s", stdout)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}

	content, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if !strings.Contains(string(content), "defer println(\"done\")\n\n\treturn") {
		t.Fatalf("expected formatted file, got:\n%s", content)
	}
}

func TestRunWithHostPathRequiresEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	mustWrite(t, path, "package sample\n")

	exitCode, _, stderr := runCLI(t, dir, "check", "--host-path", path)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stderr, engine.HostRootEnv) {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
}

func TestRunWithHostPathRejectsPositionalPaths(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	mustWrite(t, path, "package sample\n")
	t.Setenv(engine.HostRootEnv, dir)

	exitCode, _, stderr := runCLI(t, dir, "check", "--host-path", path, dir)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stderr, "cannot be used with positional paths") {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
}

func runCLI(t *testing.T, workdir string, args ...string) (int, string, string) {
	t.Helper()

	oldwd, err := os.Getwd()

	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	defer func() {
		_ = os.Chdir(oldwd)
	}()

	var stdout strings.Builder

	var stderr strings.Builder
	exitCode := run(args, &stdout, &stderr)

	return exitCode, stdout.String(), stderr.String()
}

func mustWrite(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
