package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/acnoway/litellm-go-gateway/internal/app"
	"github.com/acnoway/litellm-go-gateway/internal/config"
	"github.com/acnoway/litellm-go-gateway/internal/provider/openai"
	"github.com/acnoway/litellm-go-gateway/internal/service"
	"github.com/acnoway/litellm-go-gateway/internal/transport/httpapi"
	"github.com/go-kratos/kratos/v2/log"
)

// main 只负责依赖装配。HTTP 细节在 transport 层，模型调用在 service/provider 层，
// 因而后续增加 Anthropic 等 provider 时无需修改路由处理逻辑。
func main() {
	// 配置从环境变量加载，避免把上游密钥写进源码或镜像。
	settings := config.Load()
	if !settings.Valid() {
		slog.Error("invalid configuration")
		os.Exit(1)
	}

	// http.Client 应被复用；它的 Transport 会复用 TCP 连接。超时来自配置，
	// 防止上游异常时请求永久占用 Gateway 的 goroutine。
	client := &http.Client{Timeout: settings.OpenAI.Timeout}
	openAIProvider := openai.New(settings.OpenAI.APIKey, settings.OpenAI.BaseURL, client)

	// Handler 只处理 HTTP 协议。Service 接收 Provider 接口，形成依赖倒置，
	// 所以未来 Router 选定任意 provider 后都能复用同一业务入口。
	chatService := service.NewChatService(openAIProvider)
	handler := httpapi.NewHandler(chatService)
	router := httpapi.NewRouter(handler, settings.GatewayAPIKey)

	// Gin 负责对外 HTTP 路由；自定义 HTTPServer 实现 Kratos 的 Server 接口，
	// 使 Kratos 能统一处理进程启动、信号捕获与优雅停止。
	server := app.NewHTTPServer(settings.Address, router)
	application := app.New(server, log.NewStdLogger(os.Stdout))

	if err := application.Run(); err != nil {
		slog.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}
