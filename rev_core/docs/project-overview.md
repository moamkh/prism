# Project Overview

## Purpose

`rev_core` provides an OpenAI-compatible API gateway that lets clients call one stable endpoint while the gateway:

- authenticates client bearer tokens
- applies per-token and per-model token quotas
- forwards requests to configured upstream LLM providers
- records usage for audit visibility
- enforces per-provider concurrent request limits

## Main Capabilities

- OpenAI-compatible endpoints:
  - `/v1/responses`
  - `/v1/chat/completions`
  - `/v1/completions`
  - `/v1/models`
  - `/v1/models/{model_id}`
  - `/v1/tools`
  - `/v1/tools/execute`
- Bearer token authentication with SHA-256 hashing
- Per-token rate limiting (requests per minute)
- Per-model token limits (max input / max output tokens)
- Per-provider concurrent request limiting
- Async usage logging to PostgreSQL
- Runtime Swagger docs

## Current Repository Scope

- Active implementation: Go (`rev_core/`)
- Admin panel: Python FastAPI (`admin_pakage/admin_core/`) — local network only
- Admin UI: React Vite (`admin_pakage/admin_ui/`)

## High-Level Runtime Flow

1. Client sends request to gateway route.
2. Gateway validates caller bearer token.
3. Gateway resolves the provider for the requested model.
4. Gateway checks token permissions and applies token limits.
5. Gateway acquires a slot from the provider's concurrent limiter.
6. Gateway forwards request to upstream provider.
7. Gateway returns upstream response to the client.
8. Gateway extracts and stores usage tokens.
