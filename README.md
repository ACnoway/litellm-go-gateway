# LiteLLM Go Gateway

A Go AI Gateway scaffold inspired by LiteLLM's proxy architecture. It currently exposes an OpenAI-compatible chat endpoint and supports OpenAI as its first upstream provider.

## Included

- Gin HTTP API with `POST /v1/chat/completions` and `GET /v1/models`
- Streaming SSE passthrough and non-streaming JSON responses
- Optional gateway API key authentication
- Kratos application lifecycle and graceful HTTP shutdown
- Provider interface isolated from HTTP handlers
- OpenAI provider adapter with timeout and OpenAI-style error mapping

## Run

Set environment variables from `.env.example`, then run:

```powershell README.md
$env:OPENAI_API_KEY = "..."
$env:GATEWAY_API_KEY = "local-gateway-key"
go run ./cmd/gateway
```

The gateway listens on `http://localhost:8080` by default.

```powershell README.md
$headers = @{ Authorization = "Bearer local-gateway-key" }
$body = @{ model = "gpt-4o-mini"; messages = @(@{ role = "user"; content = "Hello" }) } | ConvertTo-Json -Depth 4
Invoke-RestMethod http://localhost:8080/v1/chat/completions -Method Post -Headers $headers -ContentType "application/json" -Body $body
```

Set `stream` to `true` in the request body to receive an SSE response.
