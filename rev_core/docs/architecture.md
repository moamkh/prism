# Architecture

## Style

The service is a small modular monolith:

- HTTP layer in `main.go` and `internal/handlers`
- provider integration layer in `internal/proxy`
- persistence access through raw SQL in `internal/db`
- token/limit enforcement in `internal/middleware`

## Directory Responsibilities

- `rev_core/main.go`: bootstraps DB, routes, middleware, docs serving, graceful shutdown.
- `rev_core/internal/proxy/proxy.go`: auth, proxy forwarding, stream handling, usage recording.
- `rev_core/internal/middleware/auth.go`: token lookup, rate limit, model permission checks, body rewriting.
- `rev_core/internal/limiter/limiter.go`: per-provider semaphore-based concurrent limiting.
- `rev_core/internal/usage/usage.go`: async batch insertion of usage logs.
- `rev_core/internal/db/db.go`: PostgreSQL connection, encryption/decryption, queries.
- `rev_core/internal/models/models.go`: Go struct mappings for DB tables.

## Request Lifecycle (Protected `/v1/*` Endpoints)

```text
Client Request
  -> HTTP route handler
  -> Auth middleware
      -> Bearer token extraction
      -> SHA-256 hash lookup in tokens table
      -> Rate limit check (requests per minute)
      -> Model extraction from request body
      -> Model permission check
      -> Token limit enforcement (rewrite max_tokens)
  -> Limiter acquire (per-provider semaphore)
  -> Reverse proxy to upstream provider
  -> Return upstream status/body/headers
  -> Limiter release
  -> Parse usage tokens from response
  -> Async usage log batch insert
```

## Provider Resolution

- If request body contains a `model` field, the gateway looks up the model UUID.
- It then finds the active provider that owns this model.
- If no model is specified, the first active provider is used.

## Proxy Transport

- Each provider can optionally have an HTTP proxy or SOCKS5 proxy.
- Proxy is enabled per-provider via the `enable_proxy` flag.
- Transport is built dynamically per request based on the resolved provider.

## Streaming Behavior

- For requests with `stream=true`, the gateway opens a streaming upstream call and pipes bytes to the client immediately.
- The gateway intercepts SSE chunks to extract usage data if present.
- If usage extraction fails after bytes are streamed, failures are silently logged.

## Documentation Exposure Model

- Public docs at `/docs` serve the full OpenAPI spec for all proxy endpoints.
- No admin endpoints are exposed on the public gateway.
