package server

// TODO struct tags
type HttpServerConfig struct {
	Address   string   `yaml:"Address"`
	TarsDirs  []string `yaml:"TarsDirs"`
	BasicAuth bool     `yaml:"BasicAuth"`
	User      string   `yaml:"User"`
	Password  string   `yaml:"Password"`
}

func DefaultConfig() *HttpServerConfig {
	return &HttpServerConfig{
		Address:   ":7600",
		BasicAuth: false,
		TarsDirs:  []string{"."},
	}
}
