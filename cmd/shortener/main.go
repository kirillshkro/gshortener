package main

import (
	"fmt"
	"log"

	"github.com/kirillshkro/gshortener/internal/app"
	"github.com/kirillshkro/gshortener/internal/config"
)

var cfg *config.Config
var buildVersion string = "N/A" // Значение по умолчанию, если не определено
var buildDate string = "N/A"    // Значение по умолчанию, если не определено
var buildCommit string = "N/A"  // Значение по умолчанию, если не определено

func main() {
	// Вывод информации о сборке
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	cfg = config.GetConfig()
	app, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
