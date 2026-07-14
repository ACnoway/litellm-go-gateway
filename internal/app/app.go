package app

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

// HTTPServer 将标准库的 http.Server 包装为 Kratos Server。
// Gin 是 http.Handler，因此无需把 Gin 路由改写为 Kratos 的 HTTP transport。
type HTTPServer struct {
	server *http.Server
}

// NewHTTPServer 构造尚未监听端口的服务实例；真正开始监听由 Kratos 调用 Start。
func NewHTTPServer(address string, handler http.Handler) *HTTPServer {
	return &HTTPServer{server: &http.Server{Addr: address, Handler: handler}}
}

// Start 会阻塞直到服务器停止。http.ErrServerClosed 是 Shutdown 的正常结果，
// 若将其向上返回，Kratos 会把一次正常关机错误地记录为启动失败。
func (s *HTTPServer) Start(context.Context) error {
	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Stop 使用传入的 context 控制优雅关机的最长时间：停止接收新连接，
// 并等待现有请求，尤其是可能持续较久的 SSE 流式请求自行完成或超时。
func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// New 创建 Kratos 应用，并把自定义 HTTPServer 注册到其生命周期管理中。
func New(server *HTTPServer, logger log.Logger) *kratos.App {
	return kratos.New(
		kratos.Name("litellm-go-gateway"),
		kratos.Server(server),
		kratos.Logger(logger),
	)
}
