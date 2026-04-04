package main

import (
	"os"
	"os/exec"
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

	if strings.TrimSpace(stderr) != "" {
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

	if strings.TrimSpace(stderr) != "" {
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

func TestRunFormatRepairsDetachedGoEmbedDirective(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")

	mustWrite(t, path, `package sample

import "embed"

//go:embed foo.txt

type runtime struct{}

var rootTemplateFS embed.FS
`)

	exitCode, stdout, stderr := runCLI(t, dir, "format", dir)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Result: fixed") {
		t.Fatalf("unexpected stdout:\n%s", stdout)
	}

	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}

	content, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if !strings.Contains(string(content), "//go:embed foo.txt\nvar rootTemplateFS embed.FS\n\ntype runtime struct{}") {
		t.Fatalf("expected go:embed directive to remain attached to the var, got:\n%s", content)
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

	if strings.TrimSpace(stderr) != "" {
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

	if strings.TrimSpace(stderr) != "" {
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

	if strings.TrimSpace(stderr) != "" {
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

func TestPrintUsage(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no args", nil},
		{"help subcommand", []string{"help"}},
		{"help flag", []string{"--help"}},
		{"short help flag", []string{"-h"}},
		{"unknown subcommand", []string{"unknown"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode, stdout, stderr := runCLI(t, t.TempDir(), tt.args...)

			// For unknown subcommand, exit code is 1. For help, it is 0. For no args, it is 1.
			expectedExitCode := 0

			if tt.name == "no args" || tt.name == "unknown subcommand" {
				expectedExitCode = 1
			}

			if exitCode != expectedExitCode {
				t.Errorf("expected exit code %d, got %d", expectedExitCode, exitCode)
			}

			if strings.TrimSpace(stdout) != "" {
				t.Errorf("expected empty stdout, got %q", stdout)
			}

			if !strings.Contains(stderr, "go-fmt check [--git-diff] [--host-path /absolute/host/path ...] [paths...]") {
				t.Errorf("expected stderr to contain usage, got %q", stderr)
			}

			if strings.Contains(stderr, "{%v}") {
				t.Errorf("stderr contains literal {%%v}: %q", stderr)
			}

			if tt.name == "unknown subcommand" && !strings.Contains(stderr, "unknown subcommand - {\"unknown\"}") {
				t.Errorf("expected stderr to contain unknown subcommand error, got %q", stderr)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	exitCode, stdout, stderr := runCLI(t, t.TempDir(), "version")

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "go-fmt dev") {
		t.Errorf("expected stdout to contain version, got %q", stdout)
	}

	if strings.TrimSpace(stderr) != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}
}

func TestRunCheckWithGitDiffFiltersTextOutput(t *testing.T) {
	dir := t.TempDir()

	mustWrite(t, filepath.Join(dir, "changed.go"), "package sample\n\nfunc run() {\nif true {\nprintln(\"ok\")\n}\nprintln(\"next\")\n}\n")
	mustWrite(t, filepath.Join(dir, "stale.go"), "package sample\n\nfunc stale() {\nif true {\nprintln(\"stale\")\n}\nprintln(\"keep\")\n}\n")

	initGitRepo(t, dir)

	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial")

	mustWrite(t, filepath.Join(dir, "changed.go"), "package sample\n\nfunc run() {\nif true {\nprintln(\"ok\")\n}\nprintln(\"next\")\nprintln(\"tail\")\n}\n")

	exitCode, stdout, stderr := runCLI(t, dir, "check", "--git-diff")

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Checked 1 file(s).") {
		t.Fatalf("expected one diff-selected file, got:\n%s", stdout)
	}

	if !strings.Contains(stdout, "changed.go") || strings.Contains(stdout, "stale.go") {
		t.Fatalf("expected only changed.go in output, got:\n%s", stdout)
	}

	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
}

func TestRunJSONWithGitDiffFiltersResults(t *testing.T) {
	dir := t.TempDir()

	mustWrite(t, filepath.Join(dir, "changed.go"), "package sample\n\nfunc run() {\nif true {\nprintln(\"ok\")\n}\nprintln(\"next\")\n}\n")
	mustWrite(t, filepath.Join(dir, "stale.go"), "package sample\n\nfunc stale() {\nif true {\nprintln(\"stale\")\n}\nprintln(\"keep\")\n}\n")

	initGitRepo(t, dir)

	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial")

	mustWrite(t, filepath.Join(dir, "changed.go"), "package sample\n\nfunc run() {\nif true {\nprintln(\"ok\")\n}\nprintln(\"next\")\nprintln(\"tail\")\n}\n")

	exitCode, stdout, stderr := runCLI(t, dir, "check", "--git-diff", "--format", "json")

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stdout, "\"files\":1") {
		t.Fatalf("expected one selected file in json output, got:\n%s", stdout)
	}

	if !strings.Contains(stdout, "changed.go") || strings.Contains(stdout, "stale.go") {
		t.Fatalf("expected json output to reference only changed.go, got:\n%s", stdout)
	}

	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
}

func TestRunAgentWithGitDiffFiltersResults(t *testing.T) {
	dir := t.TempDir()

	mustWrite(t, filepath.Join(dir, "changed.go"), "package sample\n\nfunc run() {\nif true {\nprintln(\"ok\")\n}\nprintln(\"next\")\n}\n")
	mustWrite(t, filepath.Join(dir, "stale.go"), "package sample\n\nfunc stale() {\nif true {\nprintln(\"stale\")\n}\nprintln(\"keep\")\n}\n")

	initGitRepo(t, dir)

	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial")

	mustWrite(t, filepath.Join(dir, "changed.go"), "package sample\n\nfunc run() {\nif true {\nprintln(\"ok\")\n}\nprintln(\"next\")\nprintln(\"tail\")\n}\n")

	exitCode, stdout, stderr := runCLI(t, dir, "check", "--git-diff", "--format", "agent")

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stdout, "\"files\": 1") {
		t.Fatalf("expected one selected file in agent summary, got:\n%s", stdout)
	}

	if !strings.Contains(stdout, "changed.go") || strings.Contains(stdout, "stale.go") {
		t.Fatalf("expected agent output to reference only changed.go, got:\n%s", stdout)
	}

	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
}

func TestRunCheckWithGitDiffAndHostPathsFiltersComposeMappedPaths(t *testing.T) {
	hostRoot := t.TempDir()
	workRoot := t.TempDir()
	hostAPI := filepath.Join(hostRoot, "pkg", "api")
	hostApp := filepath.Join(hostRoot, "internal", "app")

	mustWrite(t, filepath.Join(workRoot, "pkg", "api", "changed.go"), "package sample\n\nfunc run() {\nif true {\nprintln(\"ok\")\n}\nprintln(\"next\")\n}\n")
	mustWrite(t, filepath.Join(workRoot, "internal", "app", "stale.go"), "package sample\n\nfunc stale() {\nif true {\nprintln(\"stale\")\n}\nprintln(\"keep\")\n}\n")

	initGitRepo(t, workRoot)

	runGit(t, workRoot, "add", ".")
	runGit(t, workRoot, "commit", "-m", "initial")

	mustWrite(t, filepath.Join(workRoot, "pkg", "api", "changed.go"), "package sample\n\nfunc run() {\nif true {\nprintln(\"ok\")\n}\nprintln(\"next\")\nprintln(\"tail\")\n}\n")

	t.Setenv(engine.HostRootEnv, hostRoot)

	exitCode, stdout, stderr := runCLI(t, workRoot, "check", "--git-diff", "--host-path", hostAPI, "--host-path", hostApp)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Checked 1 file(s).") {
		t.Fatalf("expected one diff-selected file, got:\n%s", stdout)
	}

	if !strings.Contains(stdout, filepath.ToSlash(filepath.Join("pkg", "api", "changed.go"))) || strings.Contains(stdout, "stale.go") {
		t.Fatalf("expected only the compose-mapped changed file in output, got:\n%s", stdout)
	}

	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
}

func TestRunFormatWithGitDiffAndHostPathFormatsOnlyComposeMappedSubtree(t *testing.T) {
	hostRoot := t.TempDir()
	workRoot := t.TempDir()
	hostAPI := filepath.Join(hostRoot, "pkg", "api")
	changedPath := filepath.Join(workRoot, "pkg", "api", "changed.go")
	outsidePath := filepath.Join(workRoot, "internal", "app", "outside.go")

	mustWrite(t, changedPath, "package sample\n\nfunc run() {\nprintln(\"ok\")\n}\n")
	mustWrite(t, outsidePath, "package sample\n\nfunc run() {\nprintln(\"other\")\n}\n")

	initGitRepo(t, workRoot)

	runGit(t, workRoot, "add", ".")
	runGit(t, workRoot, "commit", "-m", "initial")

	mustWrite(t, changedPath, "package sample\n\nfunc run() {\n\tdefer println(\"done\")\n\treturn\n}\n")
	mustWrite(t, outsidePath, "package sample\n\nfunc run() {\n\tdefer println(\"skip\")\n\treturn\n}\n")

	t.Setenv(engine.HostRootEnv, hostRoot)

	exitCode, stdout, stderr := runCLI(t, workRoot, "format", "--git-diff", "--host-path", hostAPI)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Formatted 1 file(s).") {
		t.Fatalf("expected one formatted file, got:\n%s", stdout)
	}

	if strings.Contains(stdout, "outside.go") {
		t.Fatalf("expected compose target restriction to exclude outside.go, got:\n%s", stdout)
	}

	changedContent, err := os.ReadFile(changedPath)

	if err != nil {
		t.Fatalf("read changed file: %v", err)
	}

	if !strings.Contains(string(changedContent), "defer println(\"done\")\n\n\treturn") {
		t.Fatalf("expected mapped file to be formatted, got:\n%s", changedContent)
	}

	outsideContent, err := os.ReadFile(outsidePath)

	if err != nil {
		t.Fatalf("read outside file: %v", err)
	}

	if strings.Contains(string(outsideContent), "defer println(\"skip\")\n\n\treturn") {
		t.Fatalf("expected outside file to remain untouched, got:\n%s", outsideContent)
	}

	if strings.TrimSpace(stderr) != "" {
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

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "tests@example.com")
	runGit(t, dir, "config", "user.name", "Tests")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
