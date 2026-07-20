package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadReadsDotEnvWithoutOverridingProcessEnvironment(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte("GATEWAY_ADDRESS=:9090\nGATEWAY_API_KEY=file-key\nOPENAI_API_KEY=file-openai-key\nOPENAI_BASE_URL=https://example.test/v1/\nOPENAI_TIMEOUT=2m\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	keys := []string{"GATEWAY_ADDRESS", "GATEWAY_API_KEY", "OPENAI_API_KEY", "OPENAI_BASE_URL", "OPENAI_TIMEOUT"}
	originalValues := make(map[string]*string, len(keys))
	for _, key := range keys {
		if value, exists := os.LookupEnv(key); exists {
			originalValues[key] = &value
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}
	t.Cleanup(func() {
		for _, key := range keys {
			if value := originalValues[key]; value != nil {
				_ = os.Setenv(key, *value)
			} else {
				_ = os.Unsetenv(key)
			}
		}
	})

	settings := Load()
	if settings.Address != ":9090" || settings.GatewayAPIKey != "file-key" || settings.OpenAI.APIKey != "file-openai-key" || settings.OpenAI.BaseURL != "https://example.test/v1" || settings.OpenAI.Timeout != 2*time.Minute {
		t.Fatalf("Load() = %#v, want values from .env", settings)
	}

	t.Setenv("OPENAI_API_KEY", "process-key")
	if settings = Load(); settings.OpenAI.APIKey != "process-key" {
		t.Fatalf("Load().OpenAI.APIKey = %q, want process environment value", settings.OpenAI.APIKey)
	}
}
