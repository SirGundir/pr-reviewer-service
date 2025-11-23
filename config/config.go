package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		App  `yaml:"app"`
		HTTP `yaml:"http"`
		Log  `yaml:"logger"`
		PG   `yaml:"postgres"`
	}

	App struct {
		Name    string `env:"APP_NAME"    yaml:"name"    env-default:"pr-reviewer-service"`
		Version string `env:"APP_VERSION" yaml:"version" env-default:"1.0.0"`
	}

	HTTP struct {
		Port string `env:"HTTP_PORT" yaml:"port" env-default:"8080"`
	}

	Log struct {
		Level string `env:"LOG_LEVEL" yaml:"level" env-default:"info"`
	}

	PG struct {
		PoolMax int    `env:"PG_POOL_MAX" yaml:"pool_max" env-default:"2"`
		URL     string `env:"PG_URL"      yaml:"url"      env-required:"true"`
	}
)

func NewConfig() (*Config, error) {
	cfg := &Config{}

	// Read from config.yml if exists
	if _, err := os.Stat("./config/config.yml"); err == nil {
		file, err := os.ReadFile("./config/config.yml")
		if err != nil {
			return nil, fmt.Errorf("config - NewConfig - os.ReadFile: %w", err)
		}

		if err := yaml.Unmarshal(file, cfg); err != nil {
			return nil, fmt.Errorf("config - NewConfig - yaml.Unmarshal: %w", err)
		}
	}

	// Override with environment variables
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config - NewConfig - env.Parse: %w", err)
	}

	return cfg, nil
}
