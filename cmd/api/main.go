package main

import (
	"flag"
	"log"

	"github.com/anh-nguyen/resource-server/internal/infrastructure/config"
	"github.com/anh-nguyen/resource-server/internal/infrastructure/http"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	server := http.NewServer(cfg)

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}