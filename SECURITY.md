# Security Policy

## Supported Versions

Security fixes are generally applied to the latest `main` branch state.

## Reporting a Vulnerability

Please do **not** open public issues for sensitive vulnerabilities.

Instead, report security issues privately to the maintainers with:

- A clear description of the issue
- Steps to reproduce
- Expected impact
- Suggested remediation (if available)

You should receive an acknowledgement as soon as maintainers review the report.

## Security Best Practices for Deployments

- Change default `ENCRYPTION_KEY` in `rev_core/.env` before storing real provider tokens.
- Change default `SECRET_KEY` in `admin_pakage/admin_core/.env`.
- Protect `.env` files and never commit secrets to version control.
- Run behind TLS with restricted network access.
- Rotate provider API tokens and proxy access tokens periodically.
- Keep Go and Python dependencies updated.
- Monitor `/metrics` and `/status` endpoints for anomalies.
