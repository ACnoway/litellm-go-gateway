package app

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(address string, handler http.Handler) *HTTPServer {
	return &HTTPServer{server: &http.Server{Addr: address, Handler: handler}}
}

func (s *HTTPServer) Start(context.Context) error {
	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func New(server *HTTPServer, logger log.Logger) *kratos.App {
	return kratos.New(
		kratos.Name("litellm-go-gateway"),
		kratos.Server(server),
		kratos.Logger(logger),
	)
}
