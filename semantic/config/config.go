package config

type RuleToggle struct {
	Enabled bool `mapstructure:"enabled"`
}

type Rules struct {
	Spacing            RuleToggle `mapstructure:"spacing"`
	DeclarationOrder   RuleToggle `mapstructure:"declaration_order"`
	CallbackExtraction RuleToggle `mapstructure:"callback_extraction"`
	TrimSpace          RuleToggle `mapstructure:"trimspace"`
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
			Spacing:            RuleToggle{Enabled: true},
			DeclarationOrder:   RuleToggle{Enabled: true},
			CallbackExtraction: RuleToggle{Enabled: true},
			TrimSpace:          RuleToggle{Enabled: true},
		},
		Formatters: Formatters{
			Gofmt:     true,
			Goimports: true,
		},
		Exclude: []string{
			".git",
			"node_modules",
			"vendor",
		},
		NotPath: []string{},
		NotName: []string{},
	}
}
