package server

import "github.com/BurntSushi/toml"

// TODO struct tags
type HttpServerConfig struct {
	Address  string
	TarsDir  string
	User     string
	Password string
}

func ConfigFromToml(filename string) (*HttpServerConfig, error) {
	config := &HttpServerConfig{}

	_, err := toml.DecodeFile(filename, config)

	if err != nil {
		return nil, err
	}

	// TODO validate config

	return config, nil
}
