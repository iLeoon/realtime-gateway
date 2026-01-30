package config

import (
	"github.com/caarlos0/env/v10"
)

func Load() (*Config, error) {

	conf := &Config{}

	if err := env.Parse(conf); err != nil {
		return nil, err
	}

	return conf, nil

}
