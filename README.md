# Prism — OpenAI-Compatible Reverse Proxy Manager

A multi-component system to manage multiple OpenAI-compatible backends, reverse-proxy requests through a high-performance Go core, and administer everything via a Vite + React UI backed by Python FastAPI.

> **Why "Prism"?** Just like a prism splits a single beam of white light into a spectrum of colors, Prism takes a single API request and routes it to the right model across many providers — turning one stream into many possibilities.

---

## Architecture

- **rev_core** (Go, port 8080): Reverse proxy, request queueing, token authentication, usage tracking
- **admin_pakage/admin_core** (Python FastAPI, port 8000): Admin API, database models, Alembic migrations
- **admin_pakage/admin_ui** (Vite + React, port 5173): Admin dashboard
- **PostgreSQL**: Shared database for all components

---

## Prerequisites (Ubuntu)

```bash
sudo apt update
sudo apt install -y postgresql postgresql-contrib golang-go nodejs npm python3 python3-venv python3-pip
```

---

## 1. PostgreSQL Setup

```bash
sudo -u postgres psql -c "CREATE DATABASE reverse_proxy_manager_db;"
sudo -u postgres psql -c "CREATE USER admin WITH PASSWORD 'admin';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE reverse_proxy_manager_db TO admin;"
```

Update the root `.env` if needed:
```
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/reverse_proxy_manager_db
```

---

## 2. Python Admin Core Setup

```bash
cd admin_pakage/admin_core
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
alembic upgrade head
uvicorn app.main:app --host 0.0.0.0 --port 8000
```

The API will be available at `http://localhost:8000`.

---

## 3. Go Reverse Proxy Core Setup

```bash
cd rev_core
go mod tidy
go build -o rev_proxy
./rev_proxy
```

The proxy will be available at `http://localhost:8080`.

Use it as your OpenAI base URL:
```python
import openai
client = openai.OpenAI(
    api_key="rpm_your_generated_token",
    base_url="http://localhost:8080/v1"
)
```

---

## 4. React Admin UI Setup

```bash
cd admin_pakage/admin_ui
npm install
npm run dev
```

The UI will be available at `http://localhost:5173`.

To build for production:
```bash
npm run build
```

You can serve the built files with any static server, or configure Nginx:
```nginx
server {
    listen 80;
    server_name manager.yourdomain.com;
    root /path/to/admin_pakage/admin_ui/dist;
    index index.html;
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

---

## 5. Running as Systemd Services (Production)

### rev_core.service
```ini
[Unit]
Description=OpenAI Reverse Proxy Core
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/openai-proxy-server/rev_core
EnvironmentFile=/opt/openai-proxy-server/.env
EnvironmentFile=/opt/openai-proxy-server/rev_core/.env
ExecStart=/opt/openai-proxy-server/rev_core/rev_proxy
Restart=always

[Install]
WantedBy=multi-user.target
```

### admin_core.service
```ini
[Unit]
Description=OpenAI Proxy Admin API
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/openai-proxy-server/admin_pakage/admin_core
EnvironmentFile=/opt/openai-proxy-server/.env
EnvironmentFile=/opt/openai-proxy-server/admin_pakage/admin_core/.env
ExecStart=/opt/openai-proxy-server/admin_pakage/admin_core/venv/bin/uvicorn app.main:app --host 0.0.0.0 --port 8000
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable rev_core admin_core
sudo systemctl start rev_core admin_core
```

---

## 6. Nginx Reverse Proxy (Recommended)

```nginx
server {
    listen 80;
    server_name proxy.yourdomain.com;
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

server {
    listen 80;
    server_name api.yourdomain.com;
    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

server {
    listen 80;
    server_name manager.yourdomain.com;
    root /opt/openai-proxy-server/admin_pakage/admin_ui/dist;
    index index.html;
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

---

## Database Migrations

To create a new migration after changing models:
```bash
cd admin_pakage/admin_core
source venv/bin/activate
alembic revision --autogenerate -m "description"
alembic upgrade head
```

---

## Security Notes

- Change `ENCRYPTION_KEY` in `rev_core/.env` before storing real provider tokens.
- Change `SECRET_KEY` in `admin_pakage/admin_core/.env`.
- Use HTTPS in production (Let's Encrypt + Certbot).
- Provider API tokens are encrypted at rest using AES-256-GCM.
- Proxy tokens are hashed with SHA-256 for lookup.
