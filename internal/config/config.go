package config

import (
	"echo_sandbox/internal/server"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server *server.HttpServerConfig `yaml:"server"`
}

func ConfigFromToml(filename string) (*Config, error) {
	config := &Config{}

	_, err := toml.DecodeFile(filename, config)
	if err != nil {
		return nil, err
	}

	// TODO validate config

	return config, nil
}
