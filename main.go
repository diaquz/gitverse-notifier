package main

import (
	"flag"
	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/server"
)

var (
	configPath = ""
)

func init() {
	flag.StringVar(&configPath, "f", "", "config.yml path")
}

func main() {
	flag.Parse()
	config.Setup(configPath)
	logger.SetupLogger(config.GlobalConfig)

	server := server.NewHttpServer()
	logger.Fatal(server.Run())
}
