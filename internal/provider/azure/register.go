package azure

import (
	"net/http"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/config"
	"github.com/acnoway/litellm-go-gateway/internal/provider"
)

func init() {
	provider.Register(func(cfg config.Config) biz.Provider {
		if cfg.Azure.APIKey == "" || cfg.Azure.BaseURL == "" {
			return nil
		}
		client := &http.Client{Timeout: cfg.Azure.Timeout}
		return New(cfg.Azure.APIKey, cfg.Azure.BaseURL, cfg.Azure.Deployment, client)
	})
}