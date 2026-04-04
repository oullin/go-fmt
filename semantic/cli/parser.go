package cli

import (
	"flag"
	"io"
	"strings"

	"github.com/oullin/go-fmt/semantic/engine"
)

type Parser struct {
	Stderr io.Writer
}

func NewParser(stderr io.Writer) Parser {
	return Parser{Stderr: stderr}
}

type hostPathValues struct {
	paths []engine.HostPath
}

func (h *hostPathValues) String() string {
	values := make([]string, 0, len(h.paths))

	for _, path := range h.paths {
		values = append(values, string(path))
	}

	return strings.Join(values, ",")
}

func (h *hostPathValues) Set(value string) error {
	h.paths = append(h.paths, engine.HostPath(value))

	return nil
}

func (p Parser) Parse(mode Mode, args []string) (Options, error) {
	fs := flag.NewFlagSet(mode.String(), flag.ContinueOnError)
	fs.SetOutput(p.Stderr)

	configPath := fs.String("config", "", "Path to go-fmt YAML config")
	reportRoot := fs.String("cwd", "", "Path used for config discovery and report-relative file paths")
	outputFormat := fs.String("format", "text", "Output format: text, json, agent")
	gitDiff := fs.Bool("git-diff", false, "Limit the run to tracked Go files changed vs HEAD")
	var hostPaths hostPathValues
	fs.Var(&hostPaths, "host-path", "Absolute host path under HOST_PROJECT_PATH to check or format (repeatable)")

	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}

	return Options{
		Mode:         mode,
		ConfigPath:   *configPath,
		ReportRoot:   *reportRoot,
		OutputFormat: *outputFormat,
		GitDiff:      *gitDiff,
		HostPaths:    engine.HostPaths(hostPaths.paths),
		Positional:   fs.Args(),
	}, nil
}
