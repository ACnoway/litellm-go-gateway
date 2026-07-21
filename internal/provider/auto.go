package provider

import (
	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/config"
)

// Factory 是 provider 构造函数。接收完整配置，返回已实例化的 provider 或 nil（若该 provider 未配置）。
// 返回 nil 时该 provider 不会被注册到运行时 Registry。
type Factory func(config.Config) biz.Provider

var factories []Factory

// Register 在 init() 中被各 provider 包调用，将自身的构造函数加入全局列表。
// 注册顺序不重要；最终 BuildAll 会一次性实例化所有已配置的 provider。
func Register(factory Factory) {
	factories = append(factories, factory)
}

// BuildAll 遍历所有已注册的 factory，返回非 nil 的 provider 列表。
// 在 main 中调用，用于从配置自动装配所有可用 provider。
func BuildAll(cfg config.Config) []biz.Provider {
	var providers []biz.Provider
	for _, factory := range factories {
		if p := factory(cfg); p != nil {
			providers = append(providers, p)
		}
	}
	return providers
}
