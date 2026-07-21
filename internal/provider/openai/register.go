package openai

import (
	"net/http"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/config"
	"github.com/acnoway/litellm-go-gateway/internal/provider"
)

func init() {
	provider.Register(func(cfg config.Config) biz.Provider {
		// APIKey 为空时不注册此 provider，避免运行时返回 provider_not_configured
		if cfg.OpenAI.APIKey == "" {
			return nil
		}
		client := &http.Client{Timeout: cfg.OpenAI.Timeout}
		return New(cfg.OpenAI.APIKey, cfg.OpenAI.BaseURL, client)
	})
}
