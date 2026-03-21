package config

type RuleToggle struct {
	Enabled bool `mapstructure:"enabled"`
}

type Rules struct {
	Spacing RuleToggle `mapstructure:"spacing"`
}

type Formatters struct {
	Gofmt     bool `mapstructure:"gofmt"`
	Goimports bool `mapstructure:"goimports"`
}

type Config struct {
	Rules      Rules      `mapstructure:"rules"`
	Formatters Formatters `mapstructure:"formatters"`
	Exclude    []string   `mapstructure:"exclude"`
	NotPath    []string   `mapstructure:"not_path"`
	NotName    []string   `mapstructure:"not_name"`
}

func Default() Config {
	return Config{
		Rules: Rules{
			Spacing: RuleToggle{Enabled: true},
		},
		Formatters: Formatters{
			Gofmt:     true,
			Goimports: true,
		},
		Exclude: []string{
			".git",
			"vendor",
		},
		NotPath: []string{},
		NotName: []string{},
	}
}
