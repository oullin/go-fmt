package config

import (
	formatterconfig "github.com/oullin/go-fmt/packages/formatter/config"
	"github.com/oullin/go-fmt/packages/vet"
)

type Toggle struct {
	Enabled bool `mapstructure:"enabled"`
}

type Rules struct {
	Spacing Toggle `mapstructure:"spacing"`
}

type Formatters struct {
	Gofmt     bool `mapstructure:"gofmt"`
	Goimports bool `mapstructure:"goimports"`
}

type Config struct {
	Rules      Rules      `mapstructure:"rules"`
	Vet        Toggle     `mapstructure:"vet"`
	Formatters Formatters `mapstructure:"formatters"`
	Exclude    []string   `mapstructure:"exclude"`
	NotPath    []string   `mapstructure:"not_path"`
	NotName    []string   `mapstructure:"not_name"`
}

func Default() Config {
	return Config{
		Rules: Rules{
			Spacing: Toggle{Enabled: true},
		},
		Vet: Toggle{Enabled: true},
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

func (c Config) FormatterConfig() formatterconfig.Config {
	return formatterconfig.Config{
		Rules: formatterconfig.Rules{
			Spacing: formatterconfig.RuleToggle{Enabled: c.Rules.Spacing.Enabled},
		},
		Formatters: formatterconfig.Formatters{
			Gofmt:     c.Formatters.Gofmt,
			Goimports: c.Formatters.Goimports,
		},
		Exclude: c.Exclude,
		NotPath: c.NotPath,
		NotName: c.NotName,
	}
}

func (c Config) VetConfig() vet.Config {
	return vet.Config{
		Vet: vet.Toggle{Enabled: c.Vet.Enabled},
	}
}
