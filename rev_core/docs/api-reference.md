# API Reference

## Base URL

Default local base URL is `http://localhost:8080`.

## Authentication

### Bearer Token

Protected routes accept:

```
Authorization: Bearer <TOKEN>
```

The token is validated by looking up its SHA-256 hash in the `tokens` table.

## Endpoint Matrix

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| `GET` | `/health` | none | Gateway process health check. |
| `POST` | `/v1/responses` | Bearer token | Responses API passthrough. |
| `POST` | `/v1/chat/completions` | Bearer token | Chat Completions passthrough. |
| `POST` | `/v1/completions` | Bearer token | Legacy Completions passthrough. |
| `GET` | `/v1/models` | Bearer token | Lists models from upstream. |
| `GET` | `/v1/models/{model_id}` | Bearer token | Fetches a specific model payload. |
| `GET` | `/v1/tools` | Bearer token | Lists upstream tools. |
| `POST` | `/v1/tools/execute` | Bearer token | Executes upstream tool call. |

## Proxy Behavior Notes

- Request payloads are forwarded mostly unchanged.
- `max_tokens` and `max_completion_tokens` are capped according to the token's per-model limits.
- Streaming (`stream=true`) is proxied as server-sent events.
- Query parameters are forwarded as-is.

## Swagger Endpoints

- UI: `GET /docs`
- Spec JSON: `GET /swagger/doc.json`

## Error Responses

| Status | Meaning |
| --- | --- |
| `401` | Missing or invalid bearer token. |
| `403` | Token not authorized for the requested model. |
| `429` | Rate limit exceeded or too many concurrent requests for the provider. |
| `502` | No active providers available or provider for model not found. |
