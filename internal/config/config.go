package config

import (
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	Address       string
	GatewayAPIKey string
	OpenAI        OpenAIConfig
}

type OpenAIConfig struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

func Load() Config {
	return Config{
		Address:       valueOrDefault("GATEWAY_ADDRESS", ":8080"),
		GatewayAPIKey: os.Getenv("GATEWAY_API_KEY"),
		OpenAI: OpenAIConfig{
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			BaseURL: strings.TrimRight(valueOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"), "/"),
			Timeout: durationOrDefault("OPENAI_TIMEOUT", 60*time.Second),
		},
	}
}

func (c Config) Valid() bool {
	_, err := url.ParseRequestURI(c.OpenAI.BaseURL)
	return err == nil
}

func valueOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func durationOrDefault(name string, fallback time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
