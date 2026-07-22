package provider

import (
	"fmt"
	"regexp"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
)

// ModelRoute 定义一个模型到多个 provider 的映射规则
type ModelRoute struct {
	Pattern   string         // 模型名匹配模式（支持正则表达式）
	Providers []biz.Provider // 该模型对应的上游列表（按优先级排序）
}

// ModelRouter 负责根据模型名将请求路由到合适的 provider
type ModelRouter struct {
	routes        []ModelRoute
	compiledRegex []*regexp.Regexp
	fallback      biz.Provider // 默认 provider（当没有匹配的路由时使用）
}

// NewModelRouter 创建一个新的模型路由器
func NewModelRouter(fallback biz.Provider) *ModelRouter {
	return &ModelRouter{
		routes:        make([]ModelRoute, 0),
		compiledRegex: make([]*regexp.Regexp, 0),
		fallback:      fallback,
	}
}

// AddRoute 添加一个模型路由规则
// pattern 支持正则表达式，例如:
//   - "^gpt-4.*" 匹配所有 gpt-4 模型
//   - "^claude-.*" 匹配所有 claude 模型
//   - "gpt-3.5-turbo" 精确匹配
func (r *ModelRouter) AddRoute(pattern string, providers ...biz.Provider) error {
	if len(providers) == 0 {
		return fmt.Errorf("at least one provider is required for pattern %s", pattern)
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %s: %w", pattern, err)
	}

	r.routes = append(r.routes, ModelRoute{
		Pattern:   pattern,
		Providers: providers,
	})
	r.compiledRegex = append(r.compiledRegex, regex)

	return nil
}

// Route 根据模型名返回匹配的 provider 列表
// 返回第一个匹配的路由，如果没有匹配则返回 fallback provider
func (r *ModelRouter) Route(model string) []biz.Provider {
	for i, regex := range r.compiledRegex {
		if regex.MatchString(model) {
			return r.routes[i].Providers
		}
	}

	// 没有匹配的路由，使用 fallback
	if r.fallback != nil {
		return []biz.Provider{r.fallback}
	}

	return nil
}

// GetPrimaryProvider 返回指定模型的主 provider（第一个 provider）
func (r *ModelRouter) GetPrimaryProvider(model string) biz.Provider {
	providers := r.Route(model)
	if len(providers) > 0 {
		return providers[0]
	}
	return nil
}

// GetFallbackProviders 返回指定模型的 fallback providers（除第一个外的所有 providers）
func (r *ModelRouter) GetFallbackProviders(model string) []biz.Provider {
	providers := r.Route(model)
	if len(providers) > 1 {
		return providers[1:]
	}
	return nil
}
