package openai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/litellm-go-gateway/internal/biz"
)

func TestProviderChatSendsOpenAIRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		if authorization := request.Header.Get("Authorization"); authorization != "Bearer test-key" {
			t.Fatalf("Authorization = %q", authorization)
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"Hello"}]}` {
			t.Fatalf("request body = %s", body)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"id":"chatcmpl_123","object":"chat.completion"}`))
	}))
	defer server.Close()

	provider := New("test-key", server.URL+"/v1", server.Client())
	response, err := provider.Chat(context.Background(), biz.ChatRequest{
		Model:    "gpt-4o-mini",
		Messages: []biz.Message{{Role: "user", Content: "Hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(response.Body), `{"id":"chatcmpl_123","object":"chat.completion"}`; got != want {
		t.Fatalf("response body = %s, want %s", got, want)
	}
}

func TestProviderMapsOpenAIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusTooManyRequests)
		_, _ = writer.Write([]byte(`{"error":{"message":"rate limit","code":"rate_limit_exceeded"}}`))
	}))
	defer server.Close()

	provider := New("test-key", server.URL, server.Client())
	_, err := provider.Chat(context.Background(), biz.ChatRequest{
		Model:    "gpt-4o-mini",
		Messages: []biz.Message{{Role: "user", Content: "Hello"}},
	})

	providerError, ok := err.(*biz.ProviderError)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	if providerError.Status != http.StatusTooManyRequests || providerError.Code != "rate_limit_exceeded" {
		t.Fatalf("provider error = %+v", providerError)
	}
}
