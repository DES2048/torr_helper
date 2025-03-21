package config

import (
	"echo_sandbox/internal/qbt"
	"echo_sandbox/internal/server"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server *server.HttpServerConfig `yaml:"server"`
	Qbt    *qbt.QbtClientConfig     `yaml:"qbt"`
}

func ConfigFromFile(filename string) (*Config, error) {
	config := &Config{}

	err := cleanenv.ReadConfig(filename, config)
	if err != nil {
		return nil, err
	}

	// TODO: validate config

	return config, nil
}
