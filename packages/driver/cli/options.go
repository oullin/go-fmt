package cli

type Options struct {
	Mode         Mode
	ConfigPath   string
	ReportRoot   string
	OutputFormat string
	HostPath     HostPath
	Positional   []string
}
