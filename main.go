package main

import (
	"echo_sandbox/server"
	"fmt"
	"os"
	"path/filepath"

	"github.com/labstack/gommon/log"
)

var configFile string

func RunHelperServer() {
	// get config
	config, err := server.ConfigFromToml(configFile)
	if err != nil {
		log.Fatal(err)
	}

	s := server.NewHttpServer(config)
	s.Start()
}
func main() {
	// check cmd args
	if len(os.Args) <= 1 {
		fmt.Printf("Usage: %s CONFIG_FILE\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	configFile = os.Args[1]
	RunHelperServer()
}
