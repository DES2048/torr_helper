package config

import (
	"echo_sandbox/internal/qbt"
	"echo_sandbox/internal/server"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server *server.HttpServerConfig `yaml:"server"`
	Qbt    *qbt.QbtClientConfig     `yaml:"qbt"`
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
