package config

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	msgEmptyValue = "empty config value '%s'"

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
		return nil, errors.Wrap(err, "unable to read config")
	}
	cfg.WatchConfig()

	cfg.SetDefault("db.path", defaultDBPath)
	cfg.SetDefault("db.timeout", defaultDBTimeout)
	cfg.SetDefault("telegram.timeout", defaultTelegramTimeout)
	cfg.SetDefault("telegram.debug", defaultTelegramDebug)

	if err := validateConfig(cfg); err != nil {
		return nil, errors.Wrap(err, "unable to validate config")
	}

	return cfg, nil
}

func validateConfig(cfg *viper.Viper) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	for _, p := range mandatoryParams {
		if cfg.Get(p) == nil {
			return errors.Errorf(msgEmptyValue, p)
		}
	}

	if cfg.GetDuration("db.timeout") <= 0 {
		return errors.Errorf("'db.timeout' should be greater than 0")
	}

	if cfg.GetDuration("telegram.timeout") <= 0 {
		return errors.Errorf("'telegram.timeout' should be greater than 0")
	}

	return nil
}
