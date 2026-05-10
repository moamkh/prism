# Database

## Overview

The gateway uses a PostgreSQL database managed by the Python admin backend. The Go gateway connects read-only (plus usage log inserts) to the same database.

## Key Tables

| Table | Purpose |
| --- | --- |
| `providers` | Upstream backend configuration (URL, API token, proxy settings, concurrent limits). |
| `models` | Available models and their mapping to providers. |
| `tokens` | Client tokens with rate limits and global token limits. |
| `token_model_permissions` | Per-model permissions and token limits for each token. |
| `usage_logs` | Request usage records (input/output/total tokens, latency, status). |
| `config` | Global configuration key-value pairs. |

## Encryption

Provider API tokens are encrypted at rest using AES-256-GCM with a key derived via SHA-256 from `ENCRYPTION_KEY`. Both the Python admin backend and the Go gateway use the same derivation so they can read each other's data.

## Usage Logging

Usage logs are inserted in batches (up to 100 rows) every 1 second to minimize DB round-trips. If the buffer is full, new logs are dropped to avoid blocking the request path.
