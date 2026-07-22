# Model Routing Example

This example demonstrates how the model routing system automatically routes requests to the appropriate provider based on the model name.

## Setup

1. Configure multiple providers in your `.env` file:

```bash
# OpenAI provider
OPENAI_API_KEY=sk-...
OPENAI_BASE_URL=https://api.openai.com/v1

# Anthropic provider
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_BASE_URL=https://api.anthropic.com/v1

# Azure OpenAI provider (optional)
AZURE_API_KEY=...
AZURE_BASE_URL=https://your-resource.openai.azure.com
AZURE_DEPLOYMENT=gpt-4
```

2. Start the gateway:

```bash
go run ./cmd/gateway
```

## Automatic Routing

The gateway automatically routes requests based on the model name:

### GPT Models → OpenAI

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${GATEWAY_API_KEY}" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

✅ This request will be routed to the **OpenAI** provider.

### Claude Models → Anthropic

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${GATEWAY_API_KEY}" \
  -d '{
    "model": "claude-3-opus-20240229",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

✅ This request will be routed to the **Anthropic** provider.

## Fallback Example

Configure multiple providers for a model pattern to enable automatic fallback:

```go
// In your code (future: will be configurable via file)
cfg.Routing.Rules = []config.RoutingRule{
    {
        Pattern:   "^gpt-4.*",
        Providers: []string{"openai", "azure"},  // Azure as fallback
    },
}
```

### Scenario 1: Primary succeeds

```bash
# Request for gpt-4
# → Tries OpenAI first
# → ✅ OpenAI succeeds
# → Returns response from OpenAI
```

**Logs:**
```json
{"level":"INFO","msg":"starting chat completion","model":"gpt-4","primary_provider":"openai","fallback_count":1}
{"level":"INFO","msg":"chat completion succeeded","provider":"openai"}
```

### Scenario 2: Primary fails, fallback succeeds

```bash
# Request for gpt-4
# → Tries OpenAI first
# → ❌ OpenAI fails (network error)
# → Tries Azure (fallback)
# → ✅ Azure succeeds
# → Returns response from Azure
```

**Logs:**
```json
{"level":"INFO","msg":"starting chat completion","model":"gpt-4","primary_provider":"openai","fallback_count":1}
{"level":"WARN","msg":"provider failed","provider":"openai","error":"connection timeout"}
{"level":"INFO","msg":"trying fallback provider","provider":"azure","attempt":2,"total_providers":2}
{"level":"INFO","msg":"chat completion succeeded","provider":"azure"}
```

### Scenario 3: All providers fail

```bash
# Request for gpt-4
# → Tries OpenAI first
# → ❌ OpenAI fails
# → Tries Azure (fallback)
# → ❌ Azure fails
# → Returns error
```

**Logs:**
```json
{"level":"INFO","msg":"starting chat completion","model":"gpt-4","primary_provider":"openai","fallback_count":1}
{"level":"WARN","msg":"provider failed","provider":"openai","error":"connection timeout"}
{"level":"INFO","msg":"trying fallback provider","provider":"azure","attempt":2,"total_providers":2}
{"level":"WARN","msg":"provider failed","provider":"azure","error":"connection timeout"}
{"level":"ERROR","msg":"all providers failed","providers_tried":2}
```

## Monitoring

Check the usage logs to see which provider handled each request:

```bash
sqlite3 data/usage.db "SELECT request_id, model, provider, success FROM usage_logs ORDER BY created_at DESC LIMIT 10"
```

Example output:
```
abc-123|gpt-4|openai|1
def-456|claude-3-opus|anthropic|1
ghi-789|gpt-4|azure|1  ← Fallback was used
```

## Default Routing Rules

When no custom rules are configured, the gateway uses these defaults:

| Model Pattern | Provider |
|--------------|----------|
| `gpt-*` | OpenAI |
| `text-*` | OpenAI |
| `davinci*` | OpenAI |
| `claude-*` | Anthropic |
| Others | First available provider |

## Benefits

1. **Automatic failover**: If the primary provider fails, requests automatically retry with fallback providers
2. **No client changes**: Clients use standard OpenAI API format regardless of the underlying provider
3. **Transparent**: All routing decisions are logged with request IDs for debugging
4. **Flexible**: Easy to add new providers or change routing rules

## Next Steps

See [MODEL_ROUTING.md](MODEL_ROUTING.md) for:
- Advanced routing configuration
- Provider priority strategies
- Performance considerations
- Future roadmap
