package main

import (
	"github.com/aribhuiya/stormgate/internal/health_checker"
	"github.com/aribhuiya/stormgate/internal/stormgate"
	"github.com/aribhuiya/stormgate/internal/utils"
	"log"
)

func main() {
	cfg, err := utils.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	stormgateApp, err := stormgate.NewStormGate(cfg)
	if err != nil {
		log.Fatalf("Failed to create stormgateApp: %v", err)
	}
	log.Printf("\n 🌩️ Stormgate - A light weight High Performance L7 Load Balancer is starting...🚀\n Listening on %s port %d\n", cfg.Server.BindIp, cfg.Server.BindPort)

	healthCheckerService := health_checker.NewHealthCheckerService(stormgateApp.Services)
	healthCheckerService.StartService()

	stormgateApp.Serve()
}
