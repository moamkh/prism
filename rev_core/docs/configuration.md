# Configuration

## Files

- `.env`: optional runtime environment overrides.
- `rev_core/.env`: Go-gateway-specific overrides.

## Environment Variables

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `PORT` | no | `8080` | Gateway listen port. |
| `DATABASE_URL` | no | `postgresql://postgres:postgres@localhost:5432/reverse_proxy_manager_db` | PostgreSQL DSN. |
| `ENCRYPTION_KEY` | no | `changeme_32_byte_encryption_key!!` | AES-256-GCM key for encrypting provider API tokens. |

## Dotenv Load Rules

Server startup attempts to load:

1. `rev_core/.env`
2. `../.env` (project root)

## Security Recommendations

- Keep `.env` out of source control when it contains secrets.
- Use distinct provider API keys per environment.
- Treat proxy URLs as sensitive if they embed credentials.
