package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Env           string
	Port          int
	OpenAIAPIKey  string
	OpenAIModel   string
}

func Load() Config {
	// Local dev convenience. In Railway, env vars are injected, so .env won't exist.
	_ = godotenv.Load()

	cfg := Config{
		Env:          getenv("ENV", "dev"),
		Port:         mustAtoi(getenv("PORT", "8080")),
		OpenAIAPIKey: getenv("OPENAI_API_KEY", ""),
		OpenAIModel:  getenv("OPENAI_MODEL", "gpt-4.1-mini"),
	}

	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("invalid int: %q: %v", s, err)
	}
	return n
}
