package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultDBPath          = "./bolt.db"
	defaultDBTimeout       = time.Second
	defaultTelegramTimeout = 60
	defaultTelegramDebug   = false
)

var mandatoryParams = []string{
	"telegram.token",
	"commands.start",
	"commands.help",
	"messages.hello",
	"messages.help",
	"errors.unknown",
}

// New creates new viper config instance
func New(name string) (*viper.Viper, error) {
	if name == "" {
		return nil, errors.New("empty config name")
	}

	cfg := viper.New()

	cfg.SetConfigName(name)
	cfg.SetConfigType("toml")
	cfg.AddConfigPath("$HOME/.config")
	cfg.AddConfigPath("/etc")
	cfg.AddConfigPath(".")

	if err := cfg.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("unable to read config: %w", err)
	}
	cfg.WatchConfig()

	cfg.SetDefault("db.path", defaultDBPath)
	cfg.SetDefault("db.timeout", defaultDBTimeout)
	cfg.SetDefault("telegram.timeout", defaultTelegramTimeout)
	cfg.SetDefault("telegram.debug", defaultTelegramDebug)

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("unable to validate config: %w", err)
	}

	return cfg, nil
}

func validateConfig(cfg *viper.Viper) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	for _, p := range mandatoryParams {
		if cfg.Get(p) == nil {
			return fmt.Errorf("empty config value '%s'", p)
		}
	}

	if cfg.GetDuration("db.timeout") <= 0 {
		return errors.New("'db.timeout' should be greater than 0")
	}

	if cfg.GetDuration("telegram.timeout") <= 0 {
		return errors.New("'telegram.timeout' should be greater than 0")
	}

	return nil
}
