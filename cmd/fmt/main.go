package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/oullin/go-fmt/config"
	"github.com/oullin/go-fmt/engine"
)

type jsonReport struct {
	Result  string             `json:"result"`
	Files   int                `json:"files"`
	Changed int                `json:"changed"`
	Results []jsonFileResult   `json:"results,omitempty"`
	Errors  []jsonErrorMessage `json:"errors,omitempty"`
}

type jsonFileResult struct {
	File       string          `json:"file"`
	Applied    []string        `json:"applied,omitempty"`
	Violations []jsonViolation `json:"violations,omitempty"`
	Changed    bool            `json:"changed,omitempty"`
}

type jsonViolation struct {
	Rule    string `json:"rule"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

type jsonErrorMessage struct {
	File    string `json:"file"`
	Message string `json:"message"`
}

type agentReport struct {
	Result     string             `json:"result"`
	Summary    agentSummary       `json:"summary"`
	Changed    []agentChange      `json:"changed,omitempty"`
	Violations []agentViolation   `json:"violations,omitempty"`
	Errors     []jsonErrorMessage `json:"errors,omitempty"`
}

type agentSummary struct {
	Files      int `json:"files"`
	Changed    int `json:"changed"`
	Violations int `json:"violations"`
}

type agentChange struct {
	File  string   `json:"file"`
	Steps []string `json:"steps"`
}

type agentViolation struct {
	File    string `json:"file"`
	Rule    string `json:"rule"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

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
	outputFormat := fs.String("format", "text", "Output format: text, json, agent")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	cwd, err := os.Getwd()

	if err != nil {
		fmt.Fprintf(stderr, "resolve cwd: %v\n", err)

		return 1
	}

	cfg, err := config.Load(cwd, *configPath)

	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)

		return 1
	}

	runPaths := fs.Args()
	fixer := engine.New(cfg)

	var report engine.Report

	switch mode {
	case "check":
		report, err = fixer.Check(runPaths)
	case "format":
		report, err = fixer.Format(runPaths)
	}

	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)

		return 1
	}

	if err := renderReport(stdout, *outputFormat, cwd, mode, report); err != nil {
		fmt.Fprintf(stderr, "render report: %v\n", err)

		return 1
	}

	if mode == "check" {
		if report.Result == "pass" {
			return 0
		}

		return 1
	}

	if report.ErrorCount() > 0 {
		return 1
	}

	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "fmt check [paths...]")
	fmt.Fprintln(w, "fmt format [paths...]")
}

func renderReport(w io.Writer, outputFormat, cwd, mode string, report engine.Report) error {
	switch outputFormat {
	case "text":
		renderText(w, cwd, mode, report)

		return nil
	case "json":
		return json.NewEncoder(w).Encode(toJSONReport(cwd, report))
	case "agent":
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")

		return encoder.Encode(toAgentReport(cwd, report))
	default:
		return errors.New("unsupported output format")
	}
}

func renderText(w io.Writer, cwd, mode string, report engine.Report) {
	if report.Files == 0 {
		fmt.Fprintln(w, "No Go files found.")

		return
	}

	action := "Checked"

	if mode == "format" {
		action = "Formatted"
	}

	fmt.Fprintf(w, "%s %d file(s).\n", action, report.Files)

	for _, result := range report.Results {
		rel := relativePath(cwd, result.File)

		if result.Error != "" {
			fmt.Fprintf(w, "! %s %s\n", rel, result.Error)

			continue
		}

		for _, violation := range result.Violations {
			if violation.Line > 0 {
				fmt.Fprintf(w, "~ %s:%d [%s] %s\n", rel, violation.Line, violation.Rule, violation.Message)
			} else {
				fmt.Fprintf(w, "~ %s [%s] %s\n", rel, violation.Rule, violation.Message)
			}
		}

		if result.Changed {
			verb := "would apply"

			if mode == "format" {
				verb = "applied"
			}

			fmt.Fprintf(w, "  %s %s\n", verb, strings.Join(result.Applied, ", "))
		}
	}

	fmt.Fprintf(w, "Result: %s. %d changed, %d violation(s), %d error(s).\n", report.Result, report.Changed, report.ViolationCount(), report.ErrorCount())
}

func toJSONReport(cwd string, report engine.Report) jsonReport {
	out := jsonReport{
		Result:  report.Result,
		Files:   report.Files,
		Changed: report.Changed,
	}

	for _, result := range report.Results {
		rel := relativePath(cwd, result.File)

		if result.Error != "" {
			out.Errors = append(out.Errors, jsonErrorMessage{
				File:    rel,
				Message: result.Error,
			})

			continue
		}

		item := jsonFileResult{
			File:    rel,
			Applied: result.Applied,
			Changed: result.Changed,
		}

		for _, violation := range result.Violations {
			item.Violations = append(item.Violations, jsonViolation{
				Rule:    violation.Rule,
				Line:    violation.Line,
				Message: violation.Message,
			})
		}

		if item.Changed || len(item.Violations) > 0 {
			out.Results = append(out.Results, item)
		}
	}

	return out
}

func toAgentReport(cwd string, report engine.Report) agentReport {
	out := agentReport{
		Result: report.Result,
		Summary: agentSummary{
			Files:      report.Files,
			Changed:    report.Changed,
			Violations: report.ViolationCount(),
		},
	}

	for _, result := range report.Results {
		rel := relativePath(cwd, result.File)

		if result.Changed {
			out.Changed = append(out.Changed, agentChange{
				File:  rel,
				Steps: result.Applied,
			})
		}

		for _, violation := range result.Violations {
			out.Violations = append(out.Violations, agentViolation{
				File:    rel,
				Rule:    violation.Rule,
				Line:    violation.Line,
				Message: violation.Message,
			})
		}

		if result.Error != "" {
			out.Errors = append(out.Errors, jsonErrorMessage{
				File:    rel,
				Message: result.Error,
			})
		}
	}

	return out
}

func relativePath(root, path string) string {
	rel, err := filepath.Rel(root, path)

	if err != nil {
		return path
	}

	return rel
}
