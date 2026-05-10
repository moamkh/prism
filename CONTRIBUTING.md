# Contributing Guide

Thanks for contributing to the OpenAI-Compatible Reverse Proxy Manager!

## Before You Start

- Read `README.md` for architecture and setup.
- Read `CODE_OF_CONDUCT.md`.
- Check existing issues/PRs before starting large changes.

## Local Setup

### 1. PostgreSQL

```bash
sudo -u postgres psql -c "CREATE DATABASE reverse_proxy_manager_db;"
sudo -u postgres psql -c "CREATE USER admin WITH PASSWORD 'admin';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE reverse_proxy_manager_db TO admin;"
```

### 2. Python Admin Core

```bash
cd admin_pakage/admin_core
python3 -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate
pip install -r requirements.txt
alembic upgrade head
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

### 3. Go Reverse Proxy Core

```bash
cd rev_core
go mod tidy
go build -o rev_proxy
./rev_proxy  # Windows: rev_proxy.exe
```

### 4. React Admin UI

```bash
cd admin_pakage/admin_ui
npm install
npm run dev
```

### 5. Running Tests

```bash
# Python tests
cd tests
python -m pytest -xvs

# Go tests
cd rev_core
go test ./...
```

## Branch and Commit Conventions

- Create feature branches from `main`.
- Keep commits focused and small.
- Prefer conventional commit style when possible:
  - `feat: ...`
  - `fix: ...`
  - `docs: ...`
  - `refactor: ...`
  - `test: ...`
  - `chore: ...`

## Code Standards

- **Go**: Follow standard `gofmt` / `go vet`. Keep handlers concise; business logic belongs in `internal/`.
- **Python**: Keep behavior-focused docstrings. Use type hints where practical.
- **TypeScript/React**: Follow existing component patterns. Keep components modular.
- Keep comments and docstrings in English.
- Avoid mixing refactors with functional changes in one PR.

## Pull Request Checklist

- [ ] Change is scoped and described clearly.
- [ ] README / docs updated when behavior or config changes.
- [ ] DB migrations included when schema changes are introduced.
- [ ] Backward compatibility considered (API and DB).
- [ ] Tests added or updated for new logic.
- [ ] Manual test steps included in PR description.

## Areas Where Help Is Valuable

- Automated test coverage (Go unit + Python integration).
- Provider failover and retry strategies.
- Observability improvements (metrics, structured logs, tracing).
- Documentation and onboarding material.
- Security hardening and audit.
