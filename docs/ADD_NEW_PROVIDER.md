# 如何添加新的 Provider

本文档介绍如何为 litellm-go-gateway 添加新的 LLM provider（如 Anthropic、Google Gemini 等）。

## 1. 创建 Provider 实现

在 `internal/provider/<name>/` 目录下创建 `provider.go`：

```go
package anthropic

import (
    "context"
    "net/http"
    "strings"
    
    "github.com/acnoway/litellm-go-gateway/internal/biz"
)

type Provider struct {
    apiKey  string
    baseURL string
    client  *http.Client
}

func New(apiKey, baseURL string, client *http.Client) *Provider {
    return &Provider{
        apiKey:  apiKey,
        baseURL: strings.TrimRight(baseURL, "/"),
        client:  client,
    }
}

// Name 返回 provider 的唯一标识
func (p *Provider) Name() string {
    return "anthropic"
}

// Chat 实现非流式调用
func (p *Provider) Chat(ctx context.Context, req biz.ChatRequest) (biz.ChatResponse, error) {
    // 转换请求格式并调用上游 API
    // ...
}

// ChatStream 实现流式调用
func (p *Provider) ChatStream(ctx context.Context, req biz.ChatRequest) (biz.ChatStream, error) {
    // 转换请求格式并调用上游 streaming API
    // ...
}
```

## 2. 创建自动注册代码

在同一目录下创建 `register.go`：

```go
package anthropic

import (
    "net/http"
    
    "github.com/acnoway/litellm-go-gateway/internal/biz"
    "github.com/acnoway/litellm-go-gateway/internal/config"
    "github.com/acnoway/litellm-go-gateway/internal/provider"
)

func init() {
    provider.Register(func(cfg config.Config) biz.Provider {
        // 如果未配置密钥，返回 nil 跳过注册
        if cfg.Anthropic.APIKey == "" {
            return nil
        }
        client := &http.Client{Timeout: cfg.Anthropic.Timeout}
        return New(cfg.Anthropic.APIKey, cfg.Anthropic.BaseURL, client)
    })
}
```

## 3. 添加配置字段（如需要）

如果新 provider 需要自定义配置，在 `internal/config/config.go` 中添加：

```go
type Config struct {
    Address       string
    GatewayAPIKey string
    OpenAI        OpenAIConfig
    Anthropic     AnthropicConfig  // 新增
}

type AnthropicConfig struct {
    APIKey  string
    BaseURL string
    Timeout time.Duration
}

func Load() Config {
    _ = godotenv.Load()
    
    return Config{
        // ... 现有字段 ...
        Anthropic: AnthropicConfig{
            APIKey:  os.Getenv("ANTHROPIC_API_KEY"),
            BaseURL: strings.TrimRight(valueOrDefault("ANTHROPIC_BASE_URL", "https://api.anthropic.com"), "/"),
            Timeout: durationOrDefault("ANTHROPIC_TIMEOUT", 60*time.Second),
        },
    }
}
```

并在 `.env.example` 中添加示例配置：

```bash
# Anthropic Claude API
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_BASE_URL=https://api.anthropic.com
ANTHROPIC_TIMEOUT=60s
```

## 4. 在 main.go 中导入

在 `cmd/gateway/main.go` 中添加空白导入以触发注册：

```go
import (
    // ... 现有导入 ...
    _ "github.com/acnoway/litellm-go-gateway/internal/provider/openai"
    _ "github.com/acnoway/litellm-go-gateway/internal/provider/anthropic"  // 新增
)
```

## 5. 添加测试

参考 `internal/provider/openai/provider_test.go` 的模式，使用 `httptest.NewServer` 创建测试：

```go
package anthropic_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/acnoway/litellm-go-gateway/internal/biz"
    "github.com/acnoway/litellm-go-gateway/internal/provider/anthropic"
)

func TestChat(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 模拟上游响应
        w.Write([]byte(`{"content": "test response"}`))
    }))
    defer server.Close()
    
    provider := anthropic.New("test-key", server.URL, http.DefaultClient)
    resp, err := provider.Chat(context.Background(), biz.ChatRequest{
        Model: "claude-3-opus",
        Messages: []biz.Message{{Role: "user", Content: "hello"}},
    })
    
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    // ... 断言响应内容 ...
}
```

## 完成！

现在运行 `go test ./...` 验证测试通过，然后 `go run ./cmd/gateway` 启动网关。

**关键优势**：
- ✅ 无需修改 service 或 handler 层
- ✅ 无需手动在 main.go 中编写装配代码（只需一行 import）
- ✅ 未配置密钥的 provider 自动跳过，不会影响启动
- ✅ 未来可轻松扩展为基于模型名的多 provider 路由