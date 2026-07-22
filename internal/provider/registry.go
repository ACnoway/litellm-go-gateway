package provider

import (
	"fmt"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
)

// Registry 是 provider 名称到实现实例的只读索引。
// 当前 main 直接注入 OpenAI；当引入多 provider Router 时，它将成为装配入口。
type Registry struct {
	providers map[string]biz.Provider
}

// NewRegistry 在启动阶段验证 provider 名称唯一性。重复注册应当尽早失败，
// 否则运行时会静默覆盖实现，导致请求发送到非预期的上游。
func NewRegistry(providers ...biz.Provider) (*Registry, error) {
	registered := make(map[string]biz.Provider, len(providers))
	for _, item := range providers {
		if _, exists := registered[item.Name()]; exists {
			return nil, fmt.Errorf("provider %q is registered more than once", item.Name())
		}
		registered[item.Name()] = item
	}
	return &Registry{providers: registered}, nil
}

// Get 返回指定名称的 provider；第二个返回值区分”名称不存在”和”存在但值为空”。
func (r *Registry) Get(name string) (biz.Provider, bool) {
	provider, exists := r.providers[name]
	return provider, exists
}

// All 返回所有已注册的 providers
func (r *Registry) All() []biz.Provider {
	providers := make([]biz.Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	return providers
}
