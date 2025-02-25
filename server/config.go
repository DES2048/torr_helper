package server

import "github.com/BurntSushi/toml"

// TODO struct tags
type HttpServerConfig struct {
	Address   string
	TarsDir   string
	BasicAuth bool
	User      string
	Password  string
}

func DefaultConfig() *HttpServerConfig {
	return &HttpServerConfig{
		Address:   ":7600",
		BasicAuth: false,
		TarsDir:   ".",
	}
}

func ConfigFromToml(filename string) (*HttpServerConfig, error) {
	config := DefaultConfig()

	_, err := toml.DecodeFile(filename, config)

	if err != nil {
		return nil, err
	}

	// TODO validate config

	return config, nil
}
