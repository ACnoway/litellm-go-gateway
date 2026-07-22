package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config 汇集进程启动时需要的配置。当前配置只覆盖最小 OpenAI Gateway；
// 数据库、Redis、路由 deployment 等配置在对应能力实现后再加入。
type Config struct {
	Address       string
	GatewayAPIKey string
	Retry         RetryConfig
	Database      DatabaseConfig
	Routing       RoutingConfig
	OpenAI        OpenAIConfig
	Anthropic     AnthropicConfig
	Azure         AzureConfig
}

type DatabaseConfig struct {
	Path string
}

type RetryConfig struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

type RoutingConfig struct {
	Rules []RoutingRule
}

type RoutingRule struct {
	Pattern   string   // 模型名匹配模式（正则表达式）
	Providers []string // provider 名称列表（按优先级排序）
}

// OpenAIConfig 保存连接 OpenAI HTTP API 所需的信息。
// APIKey 允许为空，以便服务能启动并暴露健康检查；实际调用会返回 provider_not_configured。
type OpenAIConfig struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

type AnthropicConfig struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

type AzureConfig struct {
	APIKey     string
	BaseURL    string
	Deployment string
	Timeout    time.Duration
}

// Load 从当前工作目录的 .env 和进程环境变量建立配置，并为本地开发提供安全的非敏感默认值。
// .env 不存在或无法解析时仍会启动，以便容器和部署平台只使用进程环境变量。
// 已存在的进程环境变量不会被 .env 覆盖。
func Load() Config {
	_ = godotenv.Load()

	return Config{
		Address:       valueOrDefault("GATEWAY_ADDRESS", ":8080"),
		GatewayAPIKey: os.Getenv("GATEWAY_API_KEY"),
		Retry: RetryConfig{
			MaxAttempts:  intOrDefault("RETRY_MAX_ATTEMPTS", 3),
			InitialDelay: durationOrDefault("RETRY_INITIAL_DELAY", 100*time.Millisecond),
			MaxDelay:     durationOrDefault("RETRY_MAX_DELAY", 5*time.Second),
		},
		Database: DatabaseConfig{
			Path: valueOrDefault("DATABASE_PATH", "./data/usage.db"),
		},
		OpenAI: OpenAIConfig{
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			BaseURL: strings.TrimRight(valueOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"), "/"),
			Timeout: durationOrDefault("OPENAI_TIMEOUT", 60*time.Second),
		},
		Anthropic: AnthropicConfig{
			APIKey:  os.Getenv("ANTHROPIC_API_KEY"),
			BaseURL: strings.TrimRight(valueOrDefault("ANTHROPIC_BASE_URL", "https://api.anthropic.com/v1"), "/"),
			Timeout: durationOrDefault("ANTHROPIC_TIMEOUT", 60*time.Second),
		},
		Azure: AzureConfig{
			APIKey:     os.Getenv("AZURE_API_KEY"),
			BaseURL:    strings.TrimRight(os.Getenv("AZURE_BASE_URL"), "/"),
			Deployment: os.Getenv("AZURE_DEPLOYMENT"),
			Timeout:    durationOrDefault("AZURE_TIMEOUT", 60*time.Second),
		},
	}
}

// Valid 在监听端口前检查会导致请求构造失败的配置。
// 这里只校验 URL 格式，不访问上游网络，也不验证 API Key 的有效性。
func (c Config) Valid() bool {
	if c.OpenAI.BaseURL != "" {
		if _, err := url.ParseRequestURI(c.OpenAI.BaseURL); err != nil {
			return false
		}
	}
	if c.Anthropic.BaseURL != "" {
		if _, err := url.ParseRequestURI(c.Anthropic.BaseURL); err != nil {
			return false
		}
	}
	if c.Azure.BaseURL != "" {
		if _, err := url.ParseRequestURI(c.Azure.BaseURL); err != nil {
			return false
		}
	}
	return true
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

// intOrDefault 解析整数配置。格式不正确时回退到默认值。
func intOrDefault(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed := 0
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
		return fallback
	}
	return parsed
}
