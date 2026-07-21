package logger

import (
	"context"
	"testing"
)

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "test-request-123"

	// 将请求 ID 注入 context
	ctx = WithRequestID(ctx, requestID)

	// 从 context 中获取
	if val := ctx.Value(requestIDKey); val == nil {
		t.Fatal("request ID not found in context")
	}

	if val := ctx.Value(requestIDKey).(string); val != requestID {
		t.Errorf("expected request ID %q, got %q", requestID, val)
	}
}

func TestFromContext(t *testing.T) {
	// 初始化日志器
	Init("text")

	// 测试没有 request ID 的情况
	ctx := context.Background()
	log := FromContext(ctx)
	if log == nil {
		t.Fatal("logger should not be nil")
	}

	// 测试有 request ID 的情况
	ctx = WithRequestID(ctx, "test-request-456")
	log = FromContext(ctx)
	if log == nil {
		t.Fatal("logger should not be nil with request ID")
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"json format", "json"},
		{"text format", "text"},
		{"default format", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.format)
			if Default == nil {
				t.Fatal("Default logger should not be nil after Init")
			}
		})
	}
}
