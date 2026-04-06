package cli

import (
	"flag"
	"io"
)

type parser struct {
	stderr io.Writer
}

func newParser(stderr io.Writer) parser {
	return parser{stderr: stderr}
}

func (p parser) Parse(mode Mode, args []string) (options, error) {
	fs := flag.NewFlagSet(mode.String(), flag.ContinueOnError)
	fs.SetOutput(p.stderr)

	configPath := fs.String("config", "", "Path to go-fmt YAML config")
	reportRoot := fs.String("cwd", "", "Path used for config discovery and report-relative file paths")
	outputFormat := fs.String("format", "text", "Output format: text, json, agent")
	hostPath := fs.String("host-path", "", "Absolute host path under HOST_PROJECT_PATH to check or format")

	if err := fs.Parse(args); err != nil {
		return options{}, err
	}

	return options{
		mode:         mode,
		configPath:   *configPath,
		reportRoot:   *reportRoot,
		outputFormat: *outputFormat,
		hostPath:     HostPath(*hostPath),
		positional:   fs.Args(),
	}, nil
}
