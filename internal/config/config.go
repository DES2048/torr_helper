package config

import (
	"echo_sandbox/internal/server"

	"github.com/BurntSushi/toml"
)

func ConfigFromToml(filename string) (*server.HttpServerConfig, error) {
	config := server.DefaultConfig()

	_, err := toml.DecodeFile(filename, config)
	if err != nil {
		return nil, err
	}

	// TODO validate config

	return config, nil
}
