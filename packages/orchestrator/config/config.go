package config

import (
	"github.com/oullin/go-fmt/packages/correctness"
	semanticconfig "github.com/oullin/go-fmt/packages/semantic/config"
)

type Toggle struct {
	Enabled bool `mapstructure:"enabled"`
}

type Rules struct {
	Spacing Toggle `mapstructure:"spacing"`
}

type Correctness struct {
	Vet Toggle `mapstructure:"vet"`
}

type Formatters struct {
	Gofmt     bool `mapstructure:"gofmt"`
	Goimports bool `mapstructure:"goimports"`
}

type Config struct {
	Rules       Rules       `mapstructure:"rules"`
	Correctness Correctness `mapstructure:"correctness"`
	Formatters  Formatters  `mapstructure:"formatters"`
	Exclude     []string    `mapstructure:"exclude"`
	NotPath     []string    `mapstructure:"not_path"`
	NotName     []string    `mapstructure:"not_name"`
}

func Default() Config {
	return Config{
		Rules: Rules{
			Spacing: Toggle{Enabled: true},
		},
		Correctness: Correctness{
			Vet: Toggle{Enabled: true},
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

func (c Config) SemanticConfig() semanticconfig.Config {
	return semanticconfig.Config{
		Rules: semanticconfig.Rules{
			Spacing: semanticconfig.RuleToggle{Enabled: c.Rules.Spacing.Enabled},
		},
		Formatters: semanticconfig.Formatters{
			Gofmt:     c.Formatters.Gofmt,
			Goimports: c.Formatters.Goimports,
		},
		Exclude: c.Exclude,
		NotPath: c.NotPath,
		NotName: c.NotName,
	}
}

func (c Config) CorrectnessConfig() correctness.Config {
	return correctness.Config{
		Vet: correctness.Toggle{Enabled: c.Correctness.Vet.Enabled},
	}
}
