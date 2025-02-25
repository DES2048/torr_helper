package server

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
