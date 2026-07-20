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

Copy `.env.example` to `.env` in the project root and set the required values. The application loads this file at startup; existing process environment variables take precedence.

```powershell README.md
Copy-Item .env.example .env
# Edit .env and set OPENAI_API_KEY (and optionally GATEWAY_API_KEY).
go run ./cmd/gateway
```


The gateway listens on `http://localhost:8080` by default.

```powershell README.md
$headers = @{ Authorization = "Bearer local-gateway-key" }
$body = @{ model = "gpt-4o-mini"; messages = @(@{ role = "user"; content = "Hello" }) } | ConvertTo-Json -Depth 4
Invoke-RestMethod http://localhost:8080/v1/chat/completions -Method Post -Headers $headers -ContentType "application/json" -Body $body
```

Set `stream` to `true` in the request body to receive an SSE response.
