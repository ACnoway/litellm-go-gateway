# AGENTS.md — litellm-go-gateway

Comprehensive agent guidance for AI coding assistants working in this repository.

## Project Overview

**Goal**: A production-ready Go rewrite of the LiteLLM gateway layer, providing an OpenAI-compatible HTTP proxy to multiple upstream LLM providers with automatic routing, fallback, observability, and deployment management.

**Module**: `github.com/acnoway/litellm-go-gateway`  
**Go version**: 1.26+  
**Key dependencies**:
- Gin — HTTP routing and middleware
- Kratos v2 — application lifecycle and graceful shutdown
- godotenv — `.env` file loading with environment variable precedence
- SQLite (via `database/sql`) — usage logs, routing rules, and deployment configurations

**Current capabilities**:
- ✅ OpenAI-compatible chat completions API (`POST /v1/chat/completions`, streaming + non-streaming)
- ✅ Multiple provider support (OpenAI, Anthropic, Azure OpenAI)
- ✅ Automatic provider routing with fallback (via deployment configurations)
- ✅ Exponential backoff retry for transient network errors
- ✅ Token usage logging to SQLite with async writes
- ✅ Prometheus metrics (HTTP requests, provider calls, token usage)
- ✅ Admin API for managing deployments and routing rules
- ✅ Structured JSON logging with request ID propagation
- ✅ Gateway API key authentication with timing-safe comparison

---

## Architecture Principles

This project follows **hexagonal architecture** (ports & adapters) with strict layer isolation:

1. **Dependency Rule**: Dependencies flow inward only. Outer layers (transport, data) depend on inner layers (biz), never the reverse.
2. **Interface Segregation**: Business logic depends on interfaces (e.g., `biz.Provider`, `biz.UsageRepo`), allowing adapters to be swapped without changing the core.
3. **Single Responsibility**: Each layer has exactly one reason to change.

### Layer Responsibilities

| Layer | Purpose | May import | May NOT import |
|-------|---------|------------|----------------|
| `biz` | Domain types, interfaces (Provider, UsageRepo, DeploymentRepo) | Standard library only | Any internal package |
| `data` | Database implementations of `biz` interfaces | `biz`, `database/sql` | `provider`, `service`, `transport` |
| `provider` | Upstream provider adapters implementing `biz.Provider` | `biz`, `config`, `http.Client` | `service`, `transport`, `data` |
| `service` | Orchestration (retry, fallback, routing, usage logging) | `biz`, `config`, `provider` (Manager) | `transport`, `data` (only via `biz` interfaces) |
| `transport` | HTTP protocol (Gin handlers, middleware, routing) | `service`, `biz` (domain types only) | `provider`, `data` |
| `config` | Environment variable loading and validation | Standard library, `godotenv` | All internal packages |
| `cmd/gateway` | Dependency wiring only — no business logic | All internal packages (for assembly) | N/A (top-level) |

**Critical constraint**: `internal/biz` must remain **pure domain logic** with zero dependencies on frameworks, databases, or HTTP libraries. This keeps business rules testable and portable.

---

## File Map (Read These First)

### Entry Point
- [cmd/gateway/main.go](cmd/gateway/main.go) — Pure dependency wiring. Reads config, initializes database, registers providers, assembles services, and starts Kratos app. **Never add business logic here**.

### Domain Layer (Core)
- [internal/biz/chat.go](internal/biz/chat.go) — Domain types (`ChatRequest`, `Message`, `ChatResponse`, `ChatStream`) and the `Provider` interface. All providers implement this.
- [internal/biz/usage.go](internal/biz/usage.go) — `UsageLog` domain model and `UsageRepo` interface for token usage persistence.
- [internal/biz/deployment.go](internal/biz/deployment.go) — `Deployment` domain model and `DeploymentRepo` interface for logical-to-physical model mappings with load balancing.
- [internal/biz/admin.go](internal/biz/admin.go) — Admin API domain types (`ModelInfo`, `ProviderInfo`, `RoutingRuleRequest`, etc.) and `RoutingRuleRepo` interface.

### Data Layer (Persistence)
- [internal/data/usage.go](internal/data/usage.go) — SQLite implementation of `UsageRepo`. Creates `usage_logs` table and implements async logging.
- [internal/data/deployment.go](internal/data/deployment.go) — SQLite implementation of `DeploymentRepo`. Stores deployments with JSON-serialized `providers` and `weights` arrays.
- [internal/data/routing.go](internal/data/routing.go) — SQLite implementation of `RoutingRuleRepo`. Stores pattern-based routing rules (legacy, now superseded by deployments).

### Provider Layer (Adapters)
- [internal/provider/auto.go](internal/provider/auto.go) — Provider auto-registration system. Defines `Factory` type and `Register()` / `BuildAll()` functions for zero-config provider discovery.
- [internal/provider/registry.go](internal/provider/registry.go) — Provider registry that maps `name → biz.Provider`. Validates uniqueness at startup.
- [internal/provider/router.go](internal/provider/router.go) — `DeploymentRouter` that routes model names to provider lists using `DeploymentRepo`. Supports fallback providers.
- [internal/provider/manager.go](internal/provider/manager.go) — High-level `Manager` that orchestrates registry + router. Exposes `Chat()`, `ChatStream()`, and `GetProvidersForModel()` to the service layer.
- [internal/provider/openai/provider.go](internal/provider/openai/provider.go) — OpenAI adapter. Transforms `biz.ChatRequest` → OpenAI HTTP request, handles JSON + SSE responses, and wraps errors as `biz.ProviderError`.
- [internal/provider/openai/register.go](internal/provider/openai/register.go) — Auto-registers OpenAI provider via `init()` if `OPENAI_API_KEY` is set.
- [internal/provider/anthropic/](internal/provider/anthropic/) — Anthropic adapter (similar structure to OpenAI).
- [internal/provider/azure/](internal/provider/azure/) — Azure OpenAI adapter (similar structure to OpenAI).

### Service Layer (Orchestration)
- [internal/service/chat.go](internal/service/chat.go) — Orchestrates chat completions with retry (exponential backoff for network errors), automatic fallback (tries all providers in order), and async usage logging. **Key methods**: `Complete()` (non-streaming), `CompleteStream()` (streaming), `withRetry()`, `recordUsage()`.
- [internal/service/admin.go](internal/service/admin.go) — Admin API business logic. Lists providers, manages routing rules, validates patterns.
- [internal/service/deployment.go](internal/service/deployment.go) — Deployment CRUD operations with validation (strategy, weights, provider existence).

### Transport Layer (HTTP)
- [internal/transport/httpapi/handler.go](internal/transport/httpapi/handler.go) — OpenAI-compatible endpoints: `/healthz`, `/readyz`, `/metrics`, `/v1/models`, `/v1/chat/completions`. Handles streaming SSE with chunked transfer and flushing.
- [internal/transport/httpapi/admin_handler.go](internal/transport/httpapi/admin_handler.go) — Admin API endpoints under `/admin/providers`, `/admin/routing`, `/admin/deployments`.
- [internal/transport/httpapi/middleware.go](internal/transport/httpapi/middleware.go) — Middleware: `requestID()` (generates or forwards `X-Request-ID`), `authorize()` (bearer token auth with `subtle.ConstantTimeCompare`), `logging()` (structured logs), `metricsMiddleware()` (Prometheus).
- [internal/transport/httpapi/router.go](internal/transport/httpapi/router.go) — Gin router assembly with middleware chain: `Recovery → requestID → logging → metrics → authorize`.

### Configuration
- [internal/config/config.go](internal/config/config.go) — Loads `.env` file (if exists) + environment variables. **Process env always wins**. Provides defaults for non-sensitive values. Validates URLs before startup.

### Observability
- [internal/pkg/logger/logger.go](internal/pkg/logger/logger.go) — Structured logger (JSON or text format) with request ID propagation via context.
- [internal/pkg/metrics/metrics.go](internal/pkg/metrics/metrics.go) — Prometheus metrics: `http_requests_total`, `http_request_duration_seconds`, `provider_calls_total`, `provider_call_duration_seconds`, `token_usage`.

### Application Lifecycle
- [internal/app/app.go](internal/app/app.go) — Wraps `http.Server` as a Kratos `Server` for graceful shutdown. Handles `http.ErrServerClosed` silently.

---

## How the System Works

### 1. Startup Flow

```
main.go
├─ Load config (godotenv + env vars)
├─ Validate config (check URLs)
├─ Initialize SQLite (create tables if missing)
├─ Auto-discover providers (via init() registration)
│  ├─ OpenAI (if OPENAI_API_KEY set)
│  ├─ Anthropic (if ANTHROPIC_API_KEY set)
│  └─ Azure (if AZURE_API_KEY set)
├─ Build provider.Manager (registry + router + deployment repo)
├─ Wire services (ChatService, AdminService, DeploymentService)
├─ Wire HTTP handlers + middleware
├─ Start Kratos app (listen + graceful shutdown)
```

### 2. Request Flow (Chat Completion)

```
Client → POST /v1/chat/completions

HTTP Layer (transport/httpapi/)
├─ Middleware: requestID → logging → metrics → authorize
├─ handler.chatCompletions() binds JSON to biz.ChatRequest
├─ Checks request.Stream flag
│
├─ Non-streaming path:
│  └─ chatService.Complete(ctx, request)
│     ├─ providerManager.GetProvidersForModel(model) → [provider1, provider2, ...]
│     ├─ Try provider1.Chat(ctx, request)
│     │  ├─ withRetry() (exponential backoff for network errors)
│     │  ├─ If success → recordUsage() (async to SQLite) → return
│     │  └─ If failure → try provider2.Chat(ctx, request) → ...
│     └─ If all fail → return last error
│
└─ Streaming path:
   └─ chatService.CompleteStream(ctx, request)
      ├─ Try provider1.ChatStream(ctx, request)
      ├─ If stream starts → return stream (no retry)
      └─ If stream fails before start → try provider2 → ...

HTTP Layer
├─ Non-streaming: c.Data(200, "application/json", response.Body)
└─ Streaming: Set SSE headers → read stream body → write + flush chunks → close stream
```

**Key design choices**:
- **Retry policy**: Only network errors (timeout, connection refused, DNS failure) are retried. HTTP 4xx/5xx are NOT retried (fail fast).
- **Fallback behavior**: Non-streaming requests try all providers. Streaming requests only fallback before the stream starts (to avoid sending duplicate data to the client).
- **Usage logging**: Asynchronous with 5s timeout. Never blocks the response.
- **Error wrapping**: All provider errors become `biz.ProviderError{Status, Code, Message}`. The HTTP handler maps `Status` directly to HTTP status codes.

### 3. Deployment Routing Flow

```
Request: model="gpt-4-turbo"

DeploymentRouter.Route(model)
├─ repo.GetByName(ctx, "gpt-4-turbo")
│  └─ SELECT * FROM deployments WHERE name = ? AND enabled = 1
├─ If found:
│  ├─ deployment.Providers = ["openai", "azure"]
│  ├─ Look up each provider in registry
│  └─ Return []biz.Provider{openaiProvider, azureProvider}
└─ If not found:
   └─ Return []biz.Provider{fallbackProvider} (first available provider)

ChatService uses returned providers in order:
├─ Primary: providers[0]
└─ Fallbacks: providers[1:] (if primary fails)
```

**Deployment features**:
- Maps logical model names (e.g., `"gpt-4-turbo"`) to physical model names (e.g., `"gpt-4-turbo-2024-04-09"`).
- Supports multiple providers per model for automatic fallback.
- Load balancing strategies (future): `priority` (default), `round-robin`, `weighted`.

---

## How to Add a New Provider

Follow these steps to add support for a new upstream LLM provider (e.g., Google Gemini, Cohere, etc.).

### Step 1: Create Provider Implementation

Create `internal/provider/<name>/provider.go`:

```go
package gemini

import (
    "context"
    "net/http"
    "strings"
    
    "github.com/acnoway/litellm-go-gateway/internal/biz"
)

type Provider struct {
    apiKey  string
    baseURL string
    client  *http.Client
}

func New(apiKey, baseURL string, client *http.Client) *Provider {
    return &Provider{
        apiKey:  apiKey,
        baseURL: strings.TrimRight(baseURL, "/"),
        client:  client,
    }
}

// Name returns a stable provider identifier for registry and routing
func (p *Provider) Name() string {
    return "gemini"
}

// Chat implements non-streaming completions
func (p *Provider) Chat(ctx context.Context, req biz.ChatRequest) (biz.ChatResponse, error) {
    // 1. Check if API key is configured
    if p.apiKey == "" {
        return biz.ChatResponse{}, &biz.ProviderError{
            Status:  http.StatusServiceUnavailable,
            Code:    "provider_not_configured",
            Message: "Gemini API key is not configured",
        }
    }
    
    // 2. Transform biz.ChatRequest → Gemini format
    geminiReq := transformToGemini(req)
    body, err := json.Marshal(geminiReq)
    if err != nil {
        return biz.ChatResponse{}, fmt.Errorf("encode gemini request: %w", err)
    }
    
    // 3. Build HTTP request
    upstreamReq, err := http.NewRequestWithContext(
        ctx, 
        http.MethodPost, 
        p.baseURL+"/v1/models/"+req.Model+":generateContent", 
        bytes.NewReader(body),
    )
    if err != nil {
        return biz.ChatResponse{}, fmt.Errorf("build gemini request: %w", err)
    }
    upstreamReq.Header.Set("X-Goog-Api-Key", p.apiKey)
    upstreamReq.Header.Set("Content-Type", "application/json")
    
    // 4. Call upstream API
    resp, err := p.client.Do(upstreamReq)
    if err != nil {
        return biz.ChatResponse{}, fmt.Errorf("call gemini: %w", err)
    }
    defer resp.Body.Close()
    
    // 5. Check for HTTP errors
    if resp.StatusCode >= http.StatusBadRequest {
        return biz.ChatResponse{}, parseProviderError(resp)
    }
    
    // 6. Read response body
    body, err = io.ReadAll(resp.Body)
    if err != nil {
        return biz.ChatResponse{}, fmt.Errorf("read gemini response: %w", err)
    }
    
    // 7. Transform Gemini response → OpenAI format (for compatibility)
    openAIResp := transformToOpenAI(body)
    return biz.ChatResponse{Body: openAIResp}, nil
}

// ChatStream implements streaming completions
func (p *Provider) ChatStream(ctx context.Context, req biz.ChatRequest) (biz.ChatStream, error) {
    // Similar to Chat(), but:
    // 1. Set stream=true in request
    // 2. Return biz.ChatStream{Body: resp.Body} WITHOUT closing it
    // 3. The HTTP handler is responsible for closing the stream
    // ...
}

// parseProviderError extracts error details from upstream HTTP error responses
func parseProviderError(resp *http.Response) error {
    body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // Limit to 1 MiB
    
    var payload struct {
        Error struct {
            Code    string `json:"code"`
            Message string `json:"message"`
        } `json:"error"`
    }
    
    if json.Unmarshal(body, &payload) == nil && payload.Error.Message != "" {
        return &biz.ProviderError{
            Status:  resp.StatusCode,
            Code:    payload.Error.Code,
            Message: payload.Error.Message,
        }
    }
    
    return &biz.ProviderError{
        Status:  resp.StatusCode,
        Code:    "provider_error",
        Message: strings.TrimSpace(string(body)),
    }
}
```

**Critical rules**:
- **Request transformation**: Convert `biz.ChatRequest` to the provider's native format inside the adapter. Never pollute `biz.ChatRequest` with provider-specific fields.
- **Response transformation**: Convert the provider's response to OpenAI format (or return raw bytes if already compatible). This keeps the HTTP handler generic.
- **Error wrapping**: Always return `&biz.ProviderError` for HTTP errors. Never return raw `fmt.Errorf()` for provider failures.
- **Stream ownership**: In `ChatStream()`, return `resp.Body` WITHOUT closing it. The HTTP handler owns the stream and closes it after flushing all chunks.

### Step 2: Create Auto-Registration

Create `internal/provider/<name>/register.go`:

```go
package gemini

import (
    "net/http"
    
    "github.com/acnoway/litellm-go-gateway/internal/biz"
    "github.com/acnoway/litellm-go-gateway/internal/config"
    "github.com/acnoway/litellm-go-gateway/internal/provider"
)

func init() {
    provider.Register(func(cfg config.Config) biz.Provider {
        // If API key is not configured, return nil to skip registration
        if cfg.Gemini.APIKey == "" {
            return nil
        }
        
        client := &http.Client{Timeout: cfg.Gemini.Timeout}
        return New(cfg.Gemini.APIKey, cfg.Gemini.BaseURL, client)
    })
}
```

**How auto-registration works**:
1. The `init()` function runs before `main()`.
2. `provider.Register()` appends the factory function to a global list.
3. In [cmd/gateway/main.go](cmd/gateway/main.go), `provider.BuildAll(cfg)` calls all registered factories.
4. Factories returning `nil` are skipped (provider not configured).
5. Non-nil providers are assembled into the registry.

**Benefits**:
- ✅ Zero wiring code in `main.go` (just add one import line).
- ✅ Providers auto-disable if API keys are missing.
- ✅ No startup failure if a provider is unavailable.

### Step 3: Add Configuration (if needed)

If the provider needs custom settings, extend [internal/config/config.go](internal/config/config.go):

```go
type Config struct {
    Address       string
    GatewayAPIKey string
    Retry         RetryConfig
    Database      DatabaseConfig
    OpenAI        OpenAIConfig
    Anthropic     AnthropicConfig
    Azure         AzureConfig
    Gemini        GeminiConfig  // Add this
}

type GeminiConfig struct {
    APIKey  string
    BaseURL string
    Timeout time.Duration
}

func Load() Config {
    _ = godotenv.Load()
    
    return Config{
        // ... existing fields ...
        Gemini: GeminiConfig{
            APIKey:  os.Getenv("GEMINI_API_KEY"),
            BaseURL: strings.TrimRight(
                valueOrDefault("GEMINI_BASE_URL", "https://generativelanguage.googleapis.com"),
                "/",
            ),
            Timeout: durationOrDefault("GEMINI_TIMEOUT", 60*time.Second),
        },
    }
}
```

Add example values to `.env.example`:

```bash
# Google Gemini API
GEMINI_API_KEY=AI...
GEMINI_BASE_URL=https://generativelanguage.googleapis.com
GEMINI_TIMEOUT=60s
```

### Step 4: Import in main.go

Add a blank import to [cmd/gateway/main.go](cmd/gateway/main.go) to trigger registration:

```go
import (
    // ... existing imports ...
    _ "github.com/acnoway/litellm-go-gateway/internal/provider/openai"
    _ "github.com/acnoway/litellm-go-gateway/internal/provider/anthropic"
    _ "github.com/acnoway/litellm-go-gateway/internal/provider/azure"
    _ "github.com/acnoway/litellm-go-gateway/internal/provider/gemini"  // Add this line
)
```

### Step 5: Write Tests

Create `internal/provider/<name>/provider_test.go`:

```go
package gemini_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/acnoway/litellm-go-gateway/internal/biz"
    "github.com/acnoway/litellm-go-gateway/internal/provider/gemini"
)

func TestChat(t *testing.T) {
    // Create mock upstream server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validate request
        if r.Header.Get("X-Goog-Api-Key") != "test-key" {
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte(`{"error": {"code": "invalid_api_key", "message": "Invalid API key"}}`))
            return
        }
        
        // Return mock response
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"candidates": [{"content": {"parts": [{"text": "Hello"}]}}]}`))
    }))
    defer server.Close()
    
    // Create provider pointing to mock server
    provider := gemini.New("test-key", server.URL, http.DefaultClient)
    
    // Call Chat()
    resp, err := provider.Chat(context.Background(), biz.ChatRequest{
        Model: "gemini-pro",
        Messages: []biz.Message{{Role: "user", Content: "hello"}},
    })
    
    // Assert success
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if len(resp.Body) == 0 {
        t.Fatal("expected non-empty response body")
    }
}

func TestChatError(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusTooManyRequests)
        w.Write([]byte(`{"error": {"code": "rate_limit_exceeded", "message": "Quota exceeded"}}`))
    }))
    defer server.Close()
    
    provider := gemini.New("test-key", server.URL, http.DefaultClient)
    _, err := provider.Chat(context.Background(), biz.ChatRequest{
        Model: "gemini-pro",
        Messages: []biz.Message{{Role: "user", Content: "hello"}},
    })
    
    // Assert error is wrapped as biz.ProviderError
    var providerErr *biz.ProviderError
    if !errors.As(err, &providerErr) {
        t.Fatalf("expected biz.ProviderError, got %T", err)
    }
    if providerErr.Status != http.StatusTooManyRequests {
        t.Errorf("expected status 429, got %d", providerErr.Status)
    }
}
```

**Testing rules**:
- ✅ Use `httptest.NewServer` to mock upstream APIs (no real network calls).
- ✅ Test both success and error paths.
- ✅ Verify error wrapping (`biz.ProviderError`).
- ✅ Never require real API keys or network access.

### Step 6: Create Deployment

After the provider is registered, create a deployment via the admin API:

```bash
curl -X POST http://localhost:8080/admin/deployments \
  -H "Authorization: Bearer <GATEWAY_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gemini-pro",
    "actual_model": "gemini-1.5-pro",
    "providers": ["gemini"],
    "strategy": "priority",
    "description": "Google Gemini Pro model",
    "enabled": true
  }'
```

Now requests with `model="gemini-pro"` will route to the Gemini provider.

---

## Testing Strategy

### Unit Tests
- **Location**: `internal/provider/<name>/provider_test.go`
- **Scope**: Provider adapters only. Test request transformation, error parsing, and stream handling.
- **Mocking**: Use `httptest.NewServer` to mock upstream APIs.
- **Run**: `go test ./internal/provider/<name>/...`

### Integration Tests (future)
- **Scope**: End-to-end HTTP requests through the full stack (handler → service → provider → mock upstream).
- **Tools**: `httptest.NewServer` for the gateway itself + mock upstream servers.

### Manual Smoke Test
```bash
# Start the gateway
go run ./cmd/gateway

# Test health check
curl http://localhost:8080/healthz

# Test chat completion
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer <GATEWAY_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": false
  }'

# Test streaming
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer <GATEWAY_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  }'
```

---

## Development Guidelines

### Provider Function Signatures (MANDATORY)

Every provider MUST implement exactly these three methods:

```go
// Name returns a stable, lowercase, kebab-case identifier.
// Used for provider registry lookup and routing.
// MUST be unique across all providers.
// Examples: "openai", "anthropic", "azure", "gemini", "cohere"
func (p *Provider) Name() string

// Chat handles non-streaming chat completions.
// MUST:
// - Check if API key is configured (return ProviderError if missing)
// - Transform biz.ChatRequest to provider's native format
// - Make HTTP request with context propagation
// - Read and close response body completely
// - Transform provider response to OpenAI-compatible JSON
// - Wrap HTTP errors as biz.ProviderError
// MUST NOT:
// - Implement retry logic (handled by service layer)
// - Log request/response (handled by middleware)
// - Close response body on success if returning it
func (p *Provider) Chat(ctx context.Context, req biz.ChatRequest) (biz.ChatResponse, error)

// ChatStream handles streaming chat completions.
// MUST:
// - Check if API key is configured (return ProviderError if missing)
// - Set stream=true in upstream request
// - Return biz.ChatStream{Body: resp.Body} WITHOUT closing it
// - Transform SSE events to OpenAI format if needed
// MUST NOT:
// - Read from resp.Body (ownership transfers to handler)
// - Close resp.Body (handler closes it after streaming completes)
// - Implement retry logic (service layer handles pre-stream failures)
func (p *Provider) ChatStream(ctx context.Context, req biz.ChatRequest) (biz.ChatStream, error)
```

### Error Handling Patterns

#### Rule 1: Always wrap provider HTTP errors
```go
// ✅ CORRECT
if resp.StatusCode >= http.StatusBadRequest {
    return biz.ChatResponse{}, &biz.ProviderError{
        Status:  resp.StatusCode,
        Code:    "rate_limit_exceeded",
        Message: "Provider rate limit exceeded",
    }
}

// ❌ WRONG (loses HTTP status code)
if resp.StatusCode >= 400 {
    return biz.ChatResponse{}, fmt.Errorf("provider error: %s", resp.Status)
}
```

#### Rule 2: Network errors are NOT ProviderError
```go
// ✅ CORRECT (network error, will be retried by service layer)
resp, err := p.client.Do(req)
if err != nil {
    return biz.ChatResponse{}, fmt.Errorf("call provider: %w", err)
}

// ❌ WRONG (wrapping network error as ProviderError prevents retry)
resp, err := p.client.Do(req)
if err != nil {
    return biz.ChatResponse{}, &biz.ProviderError{
        Status:  http.StatusBadGateway,
        Code:    "network_error",
        Message: err.Error(),
    }
}
```

#### Rule 3: Check API key before making requests
```go
// ✅ CORRECT
func (p *Provider) Chat(ctx context.Context, req biz.ChatRequest) (biz.ChatResponse, error) {
    if p.apiKey == "" {
        return biz.ChatResponse{}, &biz.ProviderError{
            Status:  http.StatusServiceUnavailable,
            Code:    "provider_not_configured",
            Message: "Provider API key is not configured",
        }
    }
    // ... proceed with request
}

// ❌ WRONG (fails with cryptic network error instead of clear message)
func (p *Provider) Chat(ctx context.Context, req biz.ChatRequest) (biz.ChatResponse, error) {
    // Makes request with empty Authorization header
    req.Header.Set("Authorization", "Bearer "+p.apiKey)
}
```

#### Rule 4: Limit error body reads to prevent memory exhaustion
```go
// ✅ CORRECT
body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB limit

// ❌ WRONG (malicious upstream can send infinite error body)
body, err := io.ReadAll(resp.Body)
```

### Request Transformation Guidelines

#### Pattern 1: Direct marshaling (OpenAI-compatible providers)
```go
// biz.ChatRequest is already OpenAI-compatible
body, err := json.Marshal(req)
if err != nil {
    return biz.ChatResponse{}, fmt.Errorf("encode request: %w", err)
}
```

#### Pattern 2: Field mapping (similar but different schema)
```go
type ProviderRequest struct {
    Model      string             `json:"model"`
    Prompt     string             `json:"prompt"`      // Different field name
    MaxTokens  int                `json:"max_tokens"`
    Stream     bool               `json:"stream"`
}

func transformRequest(req biz.ChatRequest) ProviderRequest {
    // Convert messages array to single prompt string
    prompt := ""
    for _, msg := range req.Messages {
        prompt += msg.Role + ": " + msg.Content + "\n"
    }
    
    maxTokens := 4096
    if req.MaxTokens != nil {
        maxTokens = *req.MaxTokens
    }
    
    return ProviderRequest{
        Model:     req.Model,
        Prompt:    prompt,
        MaxTokens: maxTokens,
        Stream:    req.Stream,
    }
}
```

#### Pattern 3: Complex nested transformation (Anthropic-style)
```go
type AnthropicRequest struct {
    Model      string              `json:"model"`
    Messages   []AnthropicMessage  `json:"messages"`
    MaxTokens  int                 `json:"max_tokens"`
    Stream     bool                `json:"stream,omitempty"`
}

type AnthropicMessage struct {
    Role    string              `json:"role"`
    Content []AnthropicContent  `json:"content"`
}

type AnthropicContent struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

func transformToAnthropic(req biz.ChatRequest) AnthropicRequest {
    messages := make([]AnthropicMessage, len(req.Messages))
    for i, msg := range req.Messages {
        messages[i] = AnthropicMessage{
            Role: msg.Role,
            Content: []AnthropicContent{{
                Type: "text",
                Text: msg.Content,
            }},
        }
    }
    
    maxTokens := 4096
    if req.MaxTokens != nil {
        maxTokens = *req.MaxTokens
    }
    
    return AnthropicRequest{
        Model:     req.Model,
        Messages:  messages,
        MaxTokens: maxTokens,
        Stream:    req.Stream,
    }
}
```

### Response Transformation Guidelines

#### Pattern 1: Return raw bytes (already OpenAI-compatible)
```go
// OpenAI adapter
body, err := io.ReadAll(resp.Body)
if err != nil {
    return biz.ChatResponse{}, fmt.Errorf("read response: %w", err)
}
return biz.ChatResponse{Body: body}, nil  // Return as-is
```

#### Pattern 2: Transform to OpenAI format
```go
// Provider with different response schema
body, err := io.ReadAll(resp.Body)
if err != nil {
    return biz.ChatResponse{}, fmt.Errorf("read response: %w", err)
}

var providerResp ProviderResponse
if err := json.Unmarshal(body, &providerResp); err != nil {
    return biz.ChatResponse{}, fmt.Errorf("parse response: %w", err)
}

// Convert to OpenAI format
openAIResp := OpenAIResponse{
    ID:      providerResp.GenerationID,
    Object:  "chat.completion",
    Created: time.Now().Unix(),
    Model:   providerResp.Model,
    Choices: []Choice{{
        Index: 0,
        Message: Message{
            Role:    "assistant",
            Content: providerResp.Text,
        },
        FinishReason: "stop",
    }},
    Usage: Usage{
        PromptTokens:     providerResp.Tokens.Input,
        CompletionTokens: providerResp.Tokens.Output,
        TotalTokens:      providerResp.Tokens.Total,
    },
}

openAIBytes, err := json.Marshal(openAIResp)
if err != nil {
    return biz.ChatResponse{}, fmt.Errorf("encode response: %w", err)
}

return biz.ChatResponse{Body: openAIBytes}, nil
```

### Streaming Implementation Guidelines

#### Rule 1: Never close the response body in ChatStream
```go
// ✅ CORRECT
func (p *Provider) ChatStream(ctx context.Context, req biz.ChatRequest) (biz.ChatStream, error) {
    resp, err := p.do(ctx, req)  // Make HTTP request
    if err != nil {
        return biz.ChatStream{}, err
    }
    // DO NOT call resp.Body.Close() here
    return biz.ChatStream{Body: resp.Body}, nil  // Transfer ownership to handler
}

// ❌ WRONG (closes body before handler can read it)
func (p *Provider) ChatStream(ctx context.Context, req biz.ChatRequest) (biz.ChatStream, error) {
    resp, err := p.do(ctx, req)
    if err != nil {
        return biz.ChatStream{}, err
    }
    defer resp.Body.Close()  // ❌ BUG: Handler will read from closed stream
    return biz.ChatStream{Body: resp.Body}, nil
}
```

#### Rule 2: Transform SSE events if provider format differs
```go
// If provider uses OpenAI-compatible SSE, return as-is:
return biz.ChatStream{Body: resp.Body}, nil

// If provider uses different SSE format, wrap with transformer:
transformedBody := &SSETransformer{
    upstream: resp.Body,
    transform: func(line string) string {
        // Convert provider SSE event to OpenAI format
        // e.g., "data: {provider_format}" → "data: {openai_format}"
        return transformSSELine(line)
    },
}
return biz.ChatStream{Body: transformedBody}, nil
```

### Configuration Naming Conventions

```go
// Config struct field names: PascalCase
type Config struct {
    OpenAI    OpenAIConfig
    Anthropic AnthropicConfig
    Gemini    GeminiConfig
}

// Environment variables: SCREAMING_SNAKE_CASE with provider prefix
// Format: <PROVIDER>_<SETTING>
OPENAI_API_KEY=sk-...
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT=60s

ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_BASE_URL=https://api.anthropic.com/v1
ANTHROPIC_TIMEOUT=60s

GEMINI_API_KEY=AIza...
GEMINI_BASE_URL=https://generativelanguage.googleapis.com
GEMINI_TIMEOUT=60s
```

### HTTP Client Best Practices

#### Rule 1: Always set timeout
```go
// ✅ CORRECT (in register.go)
func init() {
    provider.Register(func(cfg config.Config) biz.Provider {
        if cfg.Gemini.APIKey == "" {
            return nil
        }
        client := &http.Client{Timeout: cfg.Gemini.Timeout}  // ✅ Timeout configured
        return New(cfg.Gemini.APIKey, cfg.Gemini.BaseURL, client)
    })
}

// ❌ WRONG (no timeout, requests can hang forever)
client := &http.Client{}
```

#### Rule 2: Use context for cancellation
```go
// ✅ CORRECT (request inherits context from handler)
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)

// ❌ WRONG (request ignores context, client cancellation won't work)
req, err := http.NewRequest(http.MethodPost, url, body)
```

#### Rule 3: Set all required headers
```go
// ✅ CORRECT (complete headers)
req.Header.Set("Authorization", "Bearer "+p.apiKey)
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Accept", "application/json")  // or "text/event-stream" for streaming
req.Header.Set("User-Agent", "litellm-go-gateway/1.0")

// ❌ WRONG (missing Content-Type can cause 400 errors)
req.Header.Set("Authorization", "Bearer "+p.apiKey)
```

---

## Common Tasks Reference

### Run Locally
```bash
# 1. Copy environment template
cp .env.example .env

# 2. Fill in API keys
# Edit .env and set at least one provider key:
#   OPENAI_API_KEY=sk-...
#   ANTHROPIC_API_KEY=sk-ant-...
#   AZURE_API_KEY=...

# 3. Set gateway auth (optional, disables auth if empty)
#   GATEWAY_API_KEY=your-secret-key

# 4. Run
go run ./cmd/gateway

# Server starts on http://localhost:8080
```

### Build Binary
```bash
# Build for current platform
go build -o bin/gateway ./cmd/gateway

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/gateway-linux ./cmd/gateway

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o bin/gateway.exe ./cmd/gateway
```

### Run Tests
```bash
# All tests
go test ./...

# Specific package
go test ./internal/provider/openai/...
go test ./internal/service/...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...

# Verbose output
go test -v ./...
```

### Database Operations
```bash
# View usage logs
sqlite3 ./data/usage.db "SELECT * FROM usage_logs ORDER BY created_at DESC LIMIT 10;"

# View deployments
sqlite3 ./data/usage.db "SELECT * FROM deployments WHERE enabled = 1;"

# Clear usage logs (for testing)
sqlite3 ./data/usage.db "DELETE FROM usage_logs;"

# Reset database (WARNING: deletes all data)
rm -f ./data/usage.db
```

### Admin API Examples

#### List Providers
```bash
curl http://localhost:8080/admin/providers \
  -H "Authorization: Bearer <GATEWAY_API_KEY>"
```

#### Create Deployment
```bash
curl -X POST http://localhost:8080/admin/deployments \
  -H "Authorization: Bearer <GATEWAY_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gpt-4-turbo",
    "actual_model": "gpt-4-turbo-2024-04-09",
    "providers": ["openai", "azure"],
    "strategy": "priority",
    "max_tokens": 128000,
    "description": "GPT-4 Turbo with Azure fallback",
    "enabled": true
  }'
```

#### Update Deployment
```bash
curl -X PUT http://localhost:8080/admin/deployments/1 \
  -H "Authorization: Bearer <GATEWAY_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gpt-4-turbo",
    "actual_model": "gpt-4-turbo-2024-04-09",
    "providers": ["azure", "openai"],
    "strategy": "round-robin",
    "enabled": true
  }'
```

#### Delete Deployment
```bash
curl -X DELETE http://localhost:8080/admin/deployments/1 \
  -H "Authorization: Bearer <GATEWAY_API_KEY>"
```

### Monitoring

#### Check Prometheus Metrics
```bash
curl http://localhost:8080/metrics
```

**Key metrics**:
- `http_requests_total{path, method, status}` — HTTP request count
- `http_request_duration_seconds{path, method, status}` — HTTP request latency histogram
- `provider_calls_total{provider, model, status}` — Provider call count
- `provider_call_duration_seconds{provider, model, status}` — Provider call latency histogram
- `token_usage{provider, model, type}` — Token usage (prompt/completion)
