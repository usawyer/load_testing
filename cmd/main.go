package main

import (
	"log"

	"github.com/usawyer/load_testing/configs"
	"github.com/usawyer/load_testing/internal/app"
	"github.com/usawyer/load_testing/pkg/logger"
)

func main() {
	cfg, err := configs.New()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	logger.New(cfg.Logger.Level)
	app.Run(cfg)
}
