package config

import (
	"net/url"
	"os"
	"strings"
	"time"
)

// Config 汇集进程启动时需要的配置。当前配置只覆盖最小 OpenAI Gateway；
// 数据库、Redis、路由 deployment 等配置在对应能力实现后再加入。
type Config struct {
	Address       string
	GatewayAPIKey string
	OpenAI        OpenAIConfig
}

// OpenAIConfig 保存连接 OpenAI HTTP API 所需的信息。
// APIKey 允许为空，以便服务能启动并暴露健康检查；实际调用会返回 provider_not_configured。
type OpenAIConfig struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

// Load 从环境变量建立配置，并为本地开发提供安全的非敏感默认值。
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

// Valid 在监听端口前检查会导致请求构造失败的配置。
// 这里只校验 URL 格式，不访问上游网络，也不验证 API Key 的有效性。
func (c Config) Valid() bool {
	_, err := url.ParseRequestURI(c.OpenAI.BaseURL)
	return err == nil
}

// valueOrDefault 统一处理可选字符串配置的默认值逻辑。
func valueOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

// durationOrDefault 使用 Go 标准 duration 语法，例如 "60s" 或 "2m"。
// 格式不正确时回退到默认值，避免仅因可恢复的配置错误导致启动后没有请求超时。
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
