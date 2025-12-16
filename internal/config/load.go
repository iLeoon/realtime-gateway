package config

import (
	"github.com/caarlos0/env/v10"
)

func Load() (*Config, error) {

	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil

}
