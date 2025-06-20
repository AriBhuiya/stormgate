package main

import (
	"github.com/aribhuiya/stormgate/internal/utils"
	"log"
)

func main() {
	cfg, err := utils.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("\n 🌩️ Stormgate - A light weight High Performance L7 Load Balancer is starting...🚀\n Listening on %s port %d\n", cfg.Server.BindIp, cfg.Server.BindPort)
}
