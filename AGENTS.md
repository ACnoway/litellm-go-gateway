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
internal/config/config.go    env loading + routing configuration
internal/data/usage.go       SQLite implementation of UsageRepo
internal/provider/openai/    OpenAI adapter (implements biz.Provider)
internal/provider/manager.go provider auto-discovery and routing
internal/provider/router.go  model-to-provider routing logic
internal/provider/registry.go name → provider lookup
internal/service/chat.go     orchestration (retries, routing, fallback, usage logging)
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
| Model routing rules | `internal/provider/router.go` or `internal/config/config.go` |
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

## Model Routing

The gateway automatically routes requests to the appropriate provider based on the model name:

- `gpt-*`, `text-*`, `davinci*` → OpenAI provider
- `claude-*` → Anthropic provider
- Unknown models → First available provider (fallback)

### Automatic Fallback

Each model can be configured with multiple providers (primary + fallbacks). When the primary provider fails:

1. **Non-streaming requests**: Try each provider in order until one succeeds
2. **Streaming requests**: Try fallbacks only if the stream hasn't started yet

See [docs/MODEL_ROUTING.md](docs/MODEL_ROUTING.md) and [docs/ROUTING_EXAMPLE.md](docs/ROUTING_EXAMPLE.md) for details.

## Deployment Management API

Deployments allow mapping logical model names (exposed to users) to physical provider instances with load balancing strategies.

### API Endpoints

All deployment endpoints are under `/admin/deployments` and require authentication.

#### List All Deployments
```
GET /admin/deployments
```

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "id": 1,
      "name": "gpt-4-turbo",
      "actual_model": "gpt-4-turbo-2024-04-09",
      "providers": ["openai", "azure"],
      "strategy": "priority",
      "weights": null,
      "max_tokens": 128000,
      "description": "GPT-4 Turbo with Azure fallback",
      "enabled": true,
      "created_at": "2026-07-22T10:00:00Z",
      "updated_at": "2026-07-22T10:00:00Z"
    }
  ]
}
```

#### Get Single Deployment
```
GET /admin/deployments/:id
```

**Response:** Single deployment object (same structure as list item above).

#### Create Deployment
```
POST /admin/deployments
Content-Type: application/json

{
  "name": "gpt-4-turbo",
  "actual_model": "gpt-4-turbo-2024-04-09",
  "providers": ["openai", "azure"],
  "strategy": "priority",
  "weights": null,
  "max_tokens": 128000,
  "description": "GPT-4 Turbo with Azure fallback",
  "enabled": true
}
```

**Fields:**
- `name` (required): Logical model name exposed to users
- `actual_model` (required): Physical model name sent to upstream providers
- `providers` (required): Array of provider names (min 1)
- `strategy` (optional): Load balancing strategy - `"priority"` (default), `"round-robin"`, or `"weighted"`
- `weights` (optional): Array of integers for weighted strategy (must match providers length)
- `max_tokens` (optional): Maximum token limit for this model
- `description` (optional): Human-readable description
- `enabled` (optional): Boolean, defaults to `true`

**Response:** Created deployment object with `201 Created` status.

#### Update Deployment
```
PUT /admin/deployments/:id
Content-Type: application/json

{
  "name": "gpt-4-turbo",
  "actual_model": "gpt-4-turbo-2024-04-09",
  "providers": ["openai", "azure"],
  "strategy": "round-robin",
  "max_tokens": 128000,
  "description": "Updated description",
  "enabled": true
}
```

**Response:** Updated deployment object with `200 OK` status.

#### Delete Deployment
```
DELETE /admin/deployments/:id
```

**Response:** `204 No Content` on success.

### Load Balancing Strategies

- **`priority`** (default): Always use the first provider, fallback to subsequent providers only on failure
- **`round-robin`**: Distribute requests evenly across all providers
- **`weighted`**: Distribute requests based on the weights array (requires `weights` field)

### Validation Rules

1. Strategy must be one of: `priority`, `round-robin`, `weighted`
2. For `weighted` strategy:
   - `weights` array length must match `providers` array length
   - All weights must be positive integers
3. Deployment name must be unique
4. At least one provider must be specified

