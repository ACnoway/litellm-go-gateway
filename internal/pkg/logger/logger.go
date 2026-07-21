package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// Default 是全局默认的结构化日志器，在 main 函数初始化时设置。
var Default *slog.Logger

// Init 初始化全局日志器。format 可选 "json" 或 "text"。
func Init(format string) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	Default = slog.New(handler)
	slog.SetDefault(Default)
}

// WithRequestID 将请求 ID 写入 context，供后续日志调用提取。
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// FromContext 从 context 中提取请求 ID 并返回带该字段的日志器。
// 若 context 中无请求 ID，则返回不含该字段的日志器。
func FromContext(ctx context.Context) *slog.Logger {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return Default.With("request_id", requestID)
	}
	return Default
}
