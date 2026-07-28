package main

import (
	"flag"
	"log"

	"github.com/yourname/dockersphere/internal/app"
	"github.com/yourname/dockersphere/internal/config"
)

func main() {
	cfgFile := flag.String("config", "", "config file path")
	flag.Parse()

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("create app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
