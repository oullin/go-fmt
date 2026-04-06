package cli

type options struct {
	mode         Mode
	configPath   string
	reportRoot   string
	outputFormat string
	hostPath     HostPath
	positional   []string
}
