package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/acnoway/litellm-go-gateway/internal/app"
	"github.com/acnoway/litellm-go-gateway/internal/config"
	"github.com/acnoway/litellm-go-gateway/internal/data"
	"github.com/acnoway/litellm-go-gateway/internal/pkg/logger"
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
	// 初始化结构化日志器，从环境变量读取日志格式（默认 json）
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "" {
		logFormat = "json"
	}
	logger.Init(logFormat)

	settings := config.Load()
	if !settings.Valid() {
		slog.Error("invalid configuration")
		os.Exit(1)
	}

	// 确保数据库目录存在
	dbDir := filepath.Dir(settings.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		slog.Error("failed to create database directory", "error", err)
		os.Exit(1)
	}

	// 初始化数据库
	db, err := data.InitDB(settings.Database.Path)
	if err != nil {
		slog.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	usageRepo := data.NewUsageRepo(db)
	routingRepo := data.NewRoutingRuleRepo(db)
	deploymentRepo := data.NewDeploymentRepo(db)

	providerManager, err := provider.NewManager(settings, db)
	if err != nil {
		slog.Error("provider setup failed", "error", err)
		os.Exit(1)
	}

	chatService := service.NewChatService(providerManager, settings.Retry, usageRepo)
	adminService := service.NewAdminService(providerManager, routingRepo)
	deploymentService := service.NewDeploymentService(deploymentRepo)

	handler := httpapi.NewHandler(chatService)
	adminHandler := httpapi.NewAdminHandler(adminService, deploymentService)
	router := httpapi.NewRouter(handler, adminHandler, settings.GatewayAPIKey)

	server := app.NewHTTPServer(settings.Address, router)
	application := app.New(server, log.NewStdLogger(os.Stdout))

	slog.Info("starting gateway",
		"address", settings.Address,
		"log_format", logFormat,
	)

	if err := application.Run(); err != nil {
		slog.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}