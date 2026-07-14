package provider

import (
	"fmt"

	"github.com/example/litellm-go-gateway/internal/biz"
)

type Registry struct {
	providers map[string]biz.Provider
}

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

func (r *Registry) Get(name string) (biz.Provider, bool) {
	provider, exists := r.providers[name]
	return provider, exists
}
