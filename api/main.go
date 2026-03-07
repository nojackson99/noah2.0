package main

import (
	"context"
	"log"

	"github.com/noah-jackson/noah2.0/api/internal/calendar"
	"github.com/noah-jackson/noah2.0/api/internal/config"
	"github.com/noah-jackson/noah2.0/api/internal/db"
	"github.com/noah-jackson/noah2.0/api/internal/httpserver"
	"github.com/noah-jackson/noah2.0/api/internal/httpserver/routes"
	"github.com/noah-jackson/noah2.0/api/internal/llm"
)

func main() {
	cfg := config.Load()

	log.Printf("Loaded config:")
	log.Printf("  ENV: %s", cfg.Env)
	log.Printf("  PORT: %d", cfg.Port)

	// Database
	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()

	// Google Calendar client
	calClient, err := calendar.NewClient(
		context.Background(),
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.GoogleRefreshToken,
	)
	if err != nil {
		log.Fatalf("create calendar client: %v", err)
	}

	// HTTP server
	srv := httpserver.New(cfg.Env)
	openai := llm.NewOpenAI(cfg.OpenAIAPIKey, cfg.OpenAIModel)

	routes.RegisterHealth(srv.Engine)
	routes.RegisterChat(srv.Engine, cfg, openai)
	routes.RegisterIngest(srv.Engine, calClient, database)
	routes.RegisterIngestNotes(srv.Engine, cfg.NotesDir, database, openai)

	if err := srv.Run(cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
