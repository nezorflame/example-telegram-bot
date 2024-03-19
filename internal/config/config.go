package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TelegramToken   string        `required:"true" split_words:"true"`
	TelegramTimeout time.Duration `required:"true" split_words:"true" default:"60s"`

	DBPath    string        `required:"true" split_words:"true" default:"./bolt.db"`
	DBTimeout time.Duration `required:"true" split_words:"true" default:"1s"`

	CmdStart string `required:"true" split_words:"true"`
	CmdHelp  string `required:"true" split_words:"true"`

	MsgHello      string `required:"true" split_words:"true"`
	MsgHelp       string `required:"true" split_words:"true"`
	MsgErrUnknown string `required:"true" split_words:"true"`
}

// New creates new Config instance from environment
func New(envFile string) (*Config, error) {
	// create config and set defaults
	cfg := &Config{}

	// load dotenv file first, if it's presented
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("unable to load dotenv files: %w", err)
		}
	}

	// fill config from env
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("unable to process config: %w", err)
	}

	// validate config
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("unable to validate config: %w", err)
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	if cfg.DBTimeout <= 0 {
		return errors.New("param DBTimeout should be greater than 0")
	}

	if cfg.TelegramTimeout <= 0 {
		return errors.New("param TelegramTimeout should be greater than 0")
	}

	return nil
}
