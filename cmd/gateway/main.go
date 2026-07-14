package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/example/litellm-go-gateway/internal/app"
	"github.com/example/litellm-go-gateway/internal/config"
	"github.com/example/litellm-go-gateway/internal/provider/openai"
	"github.com/example/litellm-go-gateway/internal/service"
	"github.com/example/litellm-go-gateway/internal/transport/httpapi"
	"github.com/go-kratos/kratos/v2/log"
)

func main() {
	settings := config.Load()
	if !settings.Valid() {
		slog.Error("invalid configuration")
		os.Exit(1)
	}

	client := &http.Client{Timeout: settings.OpenAI.Timeout}
	openAIProvider := openai.New(settings.OpenAI.APIKey, settings.OpenAI.BaseURL, client)
	chatService := service.NewChatService(openAIProvider)
	handler := httpapi.NewHandler(chatService)
	router := httpapi.NewRouter(handler, settings.GatewayAPIKey)
	server := app.NewHTTPServer(settings.Address, router)
	application := app.New(server, log.NewStdLogger(os.Stdout))

	if err := application.Run(); err != nil {
		slog.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}
