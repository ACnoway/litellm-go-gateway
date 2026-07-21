package main

import (
	"log/slog"
	"os"

	"github.com/acnoway/litellm-go-gateway/internal/app"
	"github.com/acnoway/litellm-go-gateway/internal/config"
	"github.com/acnoway/litellm-go-gateway/internal/provider"
	_ "github.com/acnoway/litellm-go-gateway/internal/provider/anthropic"
	_ "github.com/acnoway/litellm-go-gateway/internal/provider/azure"
	_ "github.com/acnoway/litellm-go-gateway/internal/provider/openai"
	"github.com/acnoway/litellm-go-gateway/internal/service"
	"github.com/acnoway/litellm-go-gateway/internal/transport/httpapi"
	"github.com/go-kratos/kratos/v2/log"
)

// main 只负责依赖装配。新增 provider 时无需修改此文件，
// 只需在 internal/provider/<name>/ 中实现并注册。
func main() {
	settings := config.Load()
	if !settings.Valid() {
		slog.Error("invalid configuration")
		os.Exit(1)
	}

	providerManager, err := provider.NewManager(settings)
	if err != nil {
		slog.Error("provider setup failed", "error", err)
		os.Exit(1)
	}

	chatService := service.NewChatService(providerManager)
	handler := httpapi.NewHandler(chatService)
	router := httpapi.NewRouter(handler, settings.GatewayAPIKey)

	server := app.NewHTTPServer(settings.Address, router)
	application := app.New(server, log.NewStdLogger(os.Stdout))

	if err := application.Run(); err != nil {
		slog.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}