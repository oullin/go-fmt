package cli

import "github.com/oullin/go-fmt/semantic/engine"

type Options struct {
	Mode         Mode
	ConfigPath   string
	ReportRoot   string
	OutputFormat string
	GitDiff      bool
	HostPaths    engine.HostPaths
	Positional   []string
}
