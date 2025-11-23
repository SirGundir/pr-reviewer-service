package main

import (
	"log"

	"pr-reviewer-service/config"
	"pr-reviewer-service/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run application
	app.Run(cfg)
}
