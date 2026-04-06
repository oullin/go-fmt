package cli

import (
	"flag"
	"io"
)

type Parser struct {
	Stderr io.Writer
}

func NewParser(stderr io.Writer) Parser {
	return Parser{Stderr: stderr}
}

func (p Parser) Parse(mode Mode, args []string) (Options, error) {
	fs := flag.NewFlagSet(mode.String(), flag.ContinueOnError)
	fs.SetOutput(p.Stderr)

	configPath := fs.String("config", "", "Path to go-fmt YAML config")
	reportRoot := fs.String("cwd", "", "Path used for config discovery and report-relative file paths")
	outputFormat := fs.String("format", "text", "Output format: text, json, agent")
	hostPath := fs.String("host-path", "", "Absolute host path under HOST_PROJECT_PATH to check or format")

	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}

	return Options{
		Mode:         mode,
		ConfigPath:   *configPath,
		ReportRoot:   *reportRoot,
		OutputFormat: *outputFormat,
		HostPath:     HostPath(*hostPath),
		Positional:   fs.Args(),
	}, nil
}
