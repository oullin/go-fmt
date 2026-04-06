package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const DefaultFileName = "config.yml"

func Load(cwd, explicitPath string) (Config, error) {
	cfg := Default()

	v := viper.New()
	v.SetDefault("rules.spacing.enabled", cfg.Rules.Spacing.Enabled)
	v.SetDefault("correctness.vet.enabled", cfg.Correctness.Vet.Enabled)
	v.SetDefault("formatters.gofmt", cfg.Formatters.Gofmt)
	v.SetDefault("formatters.goimports", cfg.Formatters.Goimports)
	v.SetDefault("exclude", cfg.Exclude)
	v.SetDefault("not_path", cfg.NotPath)
	v.SetDefault("not_name", cfg.NotName)

	if explicitPath != "" {
		v.SetConfigFile(explicitPath)
	} else {
		v.SetConfigFile(filepath.Join(cwd, DefaultFileName))
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError

		if explicitPath == "" && (errors.As(err, &notFound) || os.IsNotExist(err)) {
			return cfg, nil
		}

		return Config{}, fmt.Errorf("load config: %w", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	return cfg, nil
}
