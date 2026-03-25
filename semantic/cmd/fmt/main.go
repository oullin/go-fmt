package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/oullin/go-fmt/semantic/config"
	"github.com/oullin/go-fmt/semantic/engine"
	"github.com/oullin/go-fmt/semantic/formatter"
	"github.com/oullin/go-fmt/semantic/report"
	"github.com/oullin/go-fmt/semantic/rules"
	"github.com/oullin/go-fmt/semantic/rules/spacing"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)

		return 1
	}

	switch args[0] {
	case "check":
		return runCommand("check", args[1:], stdout, stderr)
	case "format":
		return runCommand("format", args[1:], stdout, stderr)
	case "version", "--version", "-version":
		fmt.Fprintf(stdout, "fmt %s\n", version)

		return 0
	case "help", "--help", "-h":
		printUsage(stdout)

		return 0
	default:
		fmt.Fprintf(stderr, "unknown subcommand %q\n\n", args[0])
		printUsage(stderr)

		return 1
	}
}

func runCommand(mode string, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet(mode, flag.ContinueOnError)
	fs.SetOutput(stderr)

	configPath := fs.String("config", "", "Path to go-fmt YAML config")
	reportRoot := fs.String("cwd", "", "Path used for config discovery and report-relative file paths")
	outputFormat := fs.String("format", "text", "Output format: text, json, agent")
	hostPath := fs.String("host-path", "", "Absolute host path under GO_FMT_HOST_ROOT to check or format")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	workRoot, err := os.Getwd()

	if err != nil {
		fmt.Fprintf(stderr, "resolve cwd: %v\n", err)

		return 1
	}

	cwd := workRoot

	if *reportRoot != "" {
		cwd = *reportRoot
	}

	cfg, err := config.Load(cwd, *configPath)

	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)

		return 1
	}

	var rr []rules.Rule

	if cfg.Rules.Spacing.Enabled {
		rr = append(rr, spacing.New())
	}

	var ff []formatter.Formatter

	if cfg.Formatters.Gofmt {
		ff = append(ff, formatter.NewGofmt())
	}

	if cfg.Formatters.Goimports {
		ff = append(ff, formatter.NewGoimports())
	}

	runPaths, err := resolveRunPaths(workRoot, *hostPath, fs.Args())

	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)

		return 1
	}

	fixer := engine.New(cfg, rr, ff)

	var result engine.Report

	switch mode {
	case "check":
		result, err = fixer.Check(runPaths)
	case "format":
		result, err = fixer.Format(runPaths)
	}

	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)

		return 1
	}

	if err := report.Render(stdout, *outputFormat, cwd, mode, result); err != nil {
		fmt.Fprintf(stderr, "render report: %v\n", err)

		return 1
	}

	if mode == "check" {
		if result.Result == "pass" {
			return 0
		}

		return 1
	}

	if result.ErrorCount() > 0 {
		return 1
	}

	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "fmt check [--host-path /absolute/host/path] [paths...]")
	fmt.Fprintln(w, "fmt format [--host-path /absolute/host/path] [paths...]")
}
