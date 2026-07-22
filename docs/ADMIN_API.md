# Admin API Documentation

This document describes the backend management API for configuring models, providers, and routing rules.

## Authentication

All admin API endpoints require authentication using the `GATEWAY_API_KEY` set in your environment:

```bash
Authorization: Bearer <GATEWAY_API_KEY>
```

## Base URL

All admin endpoints are prefixed with `/admin`.

## Provider Management

### List All Providers

Returns all registered providers and their availability status.

**Endpoint:** `GET /admin/providers`

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "name": "openai",
      "available": true,
      "type": "openai"
    },
    {
      "name": "anthropic",
      "available": true,
      "type": "anthropic"
    }
  ]
}
```

### Get Provider Details

Returns detailed information about a specific provider.

**Endpoint:** `GET /admin/providers/:name`

**Parameters:**
- `name` (path) - Provider name (e.g., "openai", "anthropic", "azure")

**Response:**
```json
{
  "name": "openai",
  "available": true,
  "type": "openai"
}
```

**Error Response (404):**
```json
{
  "error": {
    "message": "provider not found: unknown",
    "type": "not_found",
    "code": "provider_not_found"
  }
}
```

## Routing Rules Management

### List All Routing Rules

Returns all configured routing rules.

**Endpoint:** `GET /admin/routing/rules`

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "id": 1,
      "pattern": "^gpt-.*",
      "providers": ["openai"],
      "created_at": "2026-07-22T10:00:00Z",
      "updated_at": "2026-07-22T10:00:00Z"
    },
    {
      "id": 2,
      "pattern": "^claude-.*",
      "providers": ["anthropic", "openai"],
      "created_at": "2026-07-22T10:05:00Z",
      "updated_at": "2026-07-22T10:05:00Z"
    }
  ]
}
```

### Get Routing Rule Details

Returns a specific routing rule by ID.

**Endpoint:** `GET /admin/routing/rules/:id`

**Parameters:**
- `id` (path) - Rule ID (integer)

**Response:**
```json
{
  "id": 1,
  "pattern": "^gpt-.*",
  "providers": ["openai"],
  "created_at": "2026-07-22T10:00:00Z",
  "updated_at": "2026-07-22T10:00:00Z"
}
```

**Error Response (404):**
```json
{
  "error": {
    "message": "routing rule not found",
    "type": "not_found",
    "code": "rule_not_found"
  }
}
```

### Create Routing Rule

Creates a new routing rule for model-to-provider mapping.

**Endpoint:** `POST /admin/routing/rules`

**Request Body:**
```json
{
  "pattern": "^gpt-4.*",
  "providers": ["openai", "azure"]
}
```

**Fields:**
- `pattern` (required) - Regular expression pattern to match model names
- `providers` (required) - Array of provider names in priority order

**Response (201):**
```json
{
  "id": 3,
  "pattern": "^gpt-4.*",
  "providers": ["openai", "azure"],
  "created_at": "2026-07-22T11:00:00Z",
  "updated_at": "2026-07-22T11:00:00Z"
}
```

**Error Response (400):**
```json
{
  "error": {
    "message": "provider not found: unknown_provider",
    "type": "invalid_request_error",
    "code": "creation_failed"
  }
}
```

### Update Routing Rule

Updates an existing routing rule.

**Endpoint:** `PUT /admin/routing/rules/:id`

**Parameters:**
- `id` (path) - Rule ID (integer)

**Request Body:**
```json
{
  "pattern": "^gpt-4.*",
  "providers": ["azure", "openai"]
}
```

**Response:**
```json
{
  "id": 3,
  "pattern": "^gpt-4.*",
  "providers": ["azure", "openai"],
  "created_at": "2026-07-22T11:00:00Z",
  "updated_at": "2026-07-22T11:05:00Z"
}
```

### Delete Routing Rule

Deletes a routing rule.

**Endpoint:** `DELETE /admin/routing/rules/:id`

**Parameters:**
- `id` (path) - Rule ID (integer)

**Response (204):** No content

**Error Response (404):**
```json
{
  "error": {
    "message": "routing rule not found",
    "type": "not_found",
    "code": "rule_not_found"
  }
}
```

## Pattern Syntax

Routing rule patterns use regular expressions. Common examples:

- `^gpt-.*` - Matches all GPT models (gpt-4, gpt-3.5-turbo, etc.)
- `^claude-.*` - Matches all Claude models (claude-3-opus, claude-3-sonnet, etc.)
- `^gpt-4.*` - Matches only GPT-4 models
- `gpt-3.5-turbo` - Exact match for gpt-3.5-turbo

## Provider Priority and Fallback

When multiple providers are specified for a pattern, they are used in order:

1. **Primary provider** - The first provider in the list
2. **Fallback providers** - Remaining providers, tried in order if the primary fails

Example:
```json
{
  "pattern": "^claude-.*",
  "providers": ["anthropic", "openai"]
}
```

In this case:
- Requests for Claude models will first try Anthropic
- If Anthropic fails, it will fall back to OpenAI

## Example Usage

### Configure GPT-4 with Azure primary and OpenAI fallback

```bash
curl -X POST http://localhost:8080/admin/routing/rules \
  -H "Authorization: Bearer your-gateway-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "^gpt-4.*",
    "providers": ["azure", "openai"]
  }'
```

### List all configured providers

```bash
curl http://localhost:8080/admin/providers \
  -H "Authorization: Bearer your-gateway-api-key"
```

### Update a routing rule

```bash
curl -X PUT http://localhost:8080/admin/routing/rules/1 \
  -H "Authorization: Bearer your-gateway-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "^gpt-.*",
    "providers": ["openai"]
  }'
```

### Delete a routing rule

```bash
curl -X DELETE http://localhost:8080/admin/routing/rules/1 \
  -H "Authorization: Bearer your-gateway-api-key"
```

## Database Persistence

All routing rules are persisted to SQLite (`data/usage.db` by default). The rules are loaded at startup and can be modified at runtime through the admin API.

## Notes

- Routing rules are evaluated in the order they are returned from the database (by ID)
- The first matching pattern is used
- If no pattern matches, the default provider (first configured provider) is used
- Pattern changes take effect immediately without requiring a restart
- Deleting all rules will cause the system to fall back to default routing
