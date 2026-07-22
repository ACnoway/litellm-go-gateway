# AGENTS.md — litellm-go-gateway

Agent guidance for AI coding assistants working in this repository.

## Project Overview

Go rewrite of the LiteLLM gateway layer. Provides an OpenAI-compatible HTTP API (`POST /v1/chat/completions`, `GET /v1/models`) that proxies to upstream LLM providers. Current upstream: OpenAI (or any OpenAI-compatible endpoint).

Module: `github.com/acnoway/litellm-go-gateway`  
Go: 1.26+  
Key deps: Gin (HTTP), Kratos v2 (app lifecycle), godotenv (config)

## Layer Map

```
cmd/gateway/main.go          wiring only — no logic here
internal/biz/chat.go         domain types + Provider interface
internal/biz/usage.go        usage log domain model + UsageRepo interface
internal/config/config.go    env loading
internal/data/usage.go       SQLite implementation of UsageRepo
internal/provider/openai/    OpenAI adapter (implements biz.Provider)
internal/provider/manager.go provider auto-discovery and routing
internal/provider/registry.go name → provider lookup
internal/service/chat.go     orchestration (retries, routing, usage logging)
internal/transport/httpapi/  Gin router + handlers + middleware
internal/app/app.go          Kratos Server wrapper
```

## Adding a New Provider

1. Create `internal/provider/<name>/provider.go` implementing `biz.Provider`:
   ```go
   type Provider struct { ... }
   func (p *Provider) Name() string { return "<name>" }
   func (p *Provider) Chat(ctx context.Context, req biz.ChatRequest) (biz.ChatResponse, error)
   func (p *Provider) ChatStream(ctx context.Context, req biz.ChatRequest) (biz.ChatStream, error)
   ```
2. Create `internal/provider/<name>/register.go` with auto-registration:
   ```go
   func init() {
       provider.Register(func(cfg config.Config) biz.Provider {
           if cfg.<Name>.APIKey == "" {
               return nil  // skip if not configured
           }
           client := &http.Client{Timeout: cfg.<Name>.Timeout}
           return New(cfg.<Name>.APIKey, cfg.<Name>.BaseURL, client)
       })
   }
   ```
3. Add config fields to `internal/config/config.go` if the provider needs custom settings.
4. Import the provider package with `_` in `cmd/gateway/main.go` to trigger registration:
   ```go
   import _ "github.com/acnoway/litellm-go-gateway/internal/provider/<name>"
   ```
5. Add `httptest`-based tests (see `internal/provider/openai/provider_test.go` as the pattern).

**Note**: The provider will be automatically discovered and wired — no changes needed in service or handler layers.

## Error Handling Convention

Upstream errors must be wrapped as `biz.ProviderError`:
```go
return nil, &biz.ProviderError{Status: resp.StatusCode, Code: "...", Message: "..."}
```
The HTTP handler maps `ProviderError.Status` directly to the response status code.

## Streaming Convention

`ChatStream` returns `*biz.ChatStream` which wraps `io.ReadCloser`. The HTTP handler is responsible for:
- Reading chunks and flushing after each one (`c.Writer.Flush()`)
- Closing the stream body on completion or error

Ownership of the `http.Response.Body` passes all the way to the handler — do not close it in the provider.

## Auth Notes

- Auth middleware is in `internal/transport/httpapi/router.go`.
- Uses `crypto/subtle.ConstantTimeCompare` — do not replace with `==`.
- Empty `GATEWAY_API_KEY` disables auth (dev mode only, never production).
- `/healthz` and `/readyz` must remain auth-free for probe compatibility.

## Testing

```bash
go test ./...                                   # all tests
go test ./internal/provider/openai/...          # provider adapter only
go test ./internal/config/...                   # config loading only
```

Tests use `httptest.NewServer` — no live API keys required.  
Do not add tests that require real network access or real API keys.

## Common Tasks

### Run locally
```bash
cp .env.example .env   # fill in OPENAI_API_KEY
go run ./cmd/gateway
```

### Smoke test
```bash
curl -s http://localhost:8080/healthz
curl -s -H "Authorization: Bearer <GATEWAY_API_KEY>" http://localhost:8080/v1/models
```

### Build binary
```bash
go build -o bin/gateway ./cmd/gateway
```

## What Belongs Where

| Concern | Where to put it |
|---|---|
| New upstream provider | `internal/provider/<name>/` |
| Retry / fallback logic | `internal/service/chat.go` |
| New HTTP endpoint | `internal/transport/httpapi/handler.go` + `router.go` |
| New config variable | `internal/config/config.go` + `.env.example` |
| Domain types | `internal/biz/chat.go` or `internal/biz/usage.go` |
| Data access (SQLite) | `internal/data/` |
| App lifecycle (startup/shutdown hooks) | `internal/app/app.go` |

## Constraints

- Keep `cmd/gateway/main.go` as pure wiring — no business logic.
- `internal/biz` must not import any non-standard-library packages except other `biz` sub-packages.
- `internal/data` implements interfaces defined in `internal/biz` — `biz` never imports `data`.
- `internal/transport` must not import `internal/provider` or `internal/data` directly — only through `internal/service`.
- Process environment variables always take precedence over `.env` file values (enforced in `config.go`).

## Usage Logging

Token usage is automatically logged to SQLite (`data/usage.db` by default) after each chat completion:
- Request ID, provider, model, token counts (prompt/completion/total)
- Success status and error code (if failed)
- Request duration in milliseconds

Logging is asynchronous and does not block the response to the client. The database is created automatically on first startup.
