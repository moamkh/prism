# Operations

## Startup

```bash
cd rev_core
go build -o rev_proxy
./rev_proxy
```

The gateway will:
1. Load environment variables from `.env` files.
2. Connect to PostgreSQL.
3. Start the async usage log batcher.
4. Start the per-provider limiter poller (refreshes every 30s).
5. Listen for HTTP requests.

## Logs

The gateway logs to stdout/stderr:
- Startup messages (port, docs URL)
- Provider resolution info
- Proxy transport configuration
- SSE stream parsing (debug)

## Health Checks

- `GET /health` returns `{"status":"ok"}` — does not touch upstream.
- To check upstream health, use the provider's own health endpoint directly.

## Monitoring

Key metrics to watch:
- `usage_logs` table growth
- Provider `is_active` status
- Token `is_active` status
- Per-provider concurrent request saturation

## Graceful Shutdown

The gateway does not implement a custom graceful shutdown hook. In production, give the process a few seconds to finish in-flight requests before terminating.
