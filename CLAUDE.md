# Claude Code Instructions

This is a Go rewrite of the LiteLLM gateway layer — an OpenAI-compatible HTTP proxy to upstream LLM providers.

**Read [AGENTS.md](AGENTS.md) for complete agent guidance before making changes.**

## Quick Reference

- Module: `github.com/acnoway/litellm-go-gateway`
- Entry point: [cmd/gateway/main.go](cmd/gateway/main.go)
- Run: `go run ./cmd/gateway` (requires `.env` with `OPENAI_API_KEY`)
- Test: `go test ./...`

## Critical Rules

1. **Layer isolation** — `internal/biz` defines domain types and the `Provider` interface. All upstream adapters live in `internal/provider/<name>/` and implement `biz.Provider`. Handlers in `internal/transport/httpapi/` must only depend on `internal/service`, never directly on providers.

2. **Config precedence** — Process environment variables always win over `.env` file values. Do not change this.

3. **Auth timing safety** — Never replace `crypto/subtle.ConstantTimeCompare` with `==`. `/healthz` and `/readyz` must remain auth-free.

4. **Streaming ownership** — `http.Response.Body` ownership transfers through `biz.ChatStream` to the HTTP handler. Do not close it in the provider layer.

5. **Error wrapping** — Upstream errors must be wrapped as `biz.ProviderError{Status, Code, Message}`. The HTTP handler maps `Status` directly to response codes.

6. **Test isolation** — All tests use `httptest.NewServer`. Never add tests requiring real API keys or network access.

7. **Wiring only in main** — [cmd/gateway/main.go](cmd/gateway/main.go) is pure dependency wiring. No business logic belongs there.

## Common Tasks

See [AGENTS.md](AGENTS.md) for:
- Adding a new provider
- Testing patterns
- Where each concern belongs
- Import constraints between layers
