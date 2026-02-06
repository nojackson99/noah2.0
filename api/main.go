package main

import (
	"log"

	"github.com/noah-jackson/noah2.0/api/internal/config"
	// "github.com/noah-jackson/noah2.0/api/internal/db"
	"github.com/noah-jackson/noah2.0/api/internal/httpserver"
	"github.com/noah-jackson/noah2.0/api/internal/httpserver/routes"
	"github.com/noah-jackson/noah2.0/api/internal/llm"
)

func main() {
	cfg := config.Load()

	log.Printf("Loaded config:")
	log.Printf("  ENV: %s", cfg.Env)
	log.Printf("  PORT: %d", cfg.Port)
	log.Printf("  OPENAI_API_KEY: %s", cfg.OpenAIAPIKey)
	log.Printf("  OPENAI_MODEL: %s", cfg.OpenAIModel)

	srv := httpserver.New(cfg.Env)
	openai := llm.NewOpenAI(cfg.OpenAIAPIKey, cfg.OpenAIModel)

	routes.RegisterHealth(srv.Engine)
	routes.RegisterChat(srv.Engine, cfg, openai)

	if err := srv.Run(cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

