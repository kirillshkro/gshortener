package main

import (
	"log"

	"github.com/kirillshkro/gshortener/internal/app"
	"github.com/kirillshkro/gshortener/internal/config"
)

var cfg *config.Config

func main() {
	cfg = config.GetConfig()
	app, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
