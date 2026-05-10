# Reverse Proxy Gateway

The Go gateway is a reverse proxy that fronts one or more OpenAI-compatible upstream backends. It enforces bearer token authentication, applies per-token and per-model quotas, records usage, and routes requests by model.

## Core Components

- `main.go`: HTTP server entrypoint, routing, docs serving, graceful shutdown.
- `internal/proxy/proxy.go`: request relay, provider resolution, transport building, usage extraction.
- `internal/middleware/auth.go`: token validation, rate limiting, model permission checks.
- `internal/limiter/limiter.go`: per-provider concurrent request limiting.
- `internal/usage/usage.go`: async usage log batching.
- `internal/db/db.go`: PostgreSQL connection, provider/model/token queries.
- `docs/swagger.json`: embedded OpenAPI specification.

## Local Setup

### Requirements

- Go 1.23+
- PostgreSQL (running `reverse_proxy_manager_db`)

### Run Steps

```bash
cd rev_core
go mod tidy
go build -o rev_proxy
./rev_proxy
```

Gateway listens on `http://localhost:8080` by default.

## Environment Variables

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | HTTP listen port. |
| `DATABASE_URL` | `postgresql://postgres:postgres@localhost:5432/reverse_proxy_manager_db` | PostgreSQL DSN. |
| `ENCRYPTION_KEY` | `changeme_32_byte_encryption_key!!` | AES key for provider API token encryption. |

## HTTP API Summary

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| `GET` | `/health` | none | Gateway process health check. |
| `POST` | `/v1/responses` | Bearer token | Responses API passthrough. |
| `POST` | `/v1/chat/completions` | Bearer token | Chat Completions passthrough. |
| `POST` | `/v1/completions` | Bearer token | Legacy Completions passthrough. |
| `GET` | `/v1/models` | Bearer token | Lists models from upstream. |
| `GET` | `/v1/models/{model_id}` | Bearer token | Fetches one model metadata payload. |
| `GET` | `/v1/tools` | Bearer token | Lists upstream tools. |
| `POST` | `/v1/tools/execute` | Bearer token | Executes upstream tool call. |

## Swagger and OpenAPI

- UI: `GET /docs`
- Spec: `GET /swagger/doc.json`

## Additional Documentation

- `project-overview.md`
- `architecture.md`
- `api-reference.md`
- `configuration.md`
- `database.md`
- `operations.md`
