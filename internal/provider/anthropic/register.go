package anthropic

import (
	"net/http"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/config"
	"github.com/acnoway/litellm-go-gateway/internal/provider"
)

func init() {
	provider.Register(func(cfg config.Config) biz.Provider {
		if cfg.Anthropic.APIKey == "" {
			return nil
		}
		client := &http.Client{Timeout: cfg.Anthropic.Timeout}
		return New(cfg.Anthropic.APIKey, cfg.Anthropic.BaseURL, client)
	})
}
