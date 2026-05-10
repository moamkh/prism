"""Tests for Python admin_core API endpoints."""
import pytest


class TestProviders:
    """Test provider CRUD and proxy settings."""

    def test_create_provider_without_proxy(self, client, clean_tables):
        resp = client.post("/api/v1/providers/", json={
            "name": "TestProvider",
            "base_url": "https://api.openai.com",
            "api_token": "sk-test",
            "is_active": True,
        })
        assert resp.status_code == 200
        data = resp.json()
        assert data["name"] == "TestProvider"
        assert data["base_url"] == "https://api.openai.com"
        assert data["max_concurrent_requests"] == 100
        assert data["enable_proxy"] is True
        assert data["http_proxy"] is None
        assert data["socks5_proxy"] is None

    def test_create_provider_with_http_proxy(self, client, clean_tables):
        resp = client.post("/api/v1/providers/", json={
            "name": "ProxyProvider",
            "base_url": "https://api.openai.com",
            "api_token": "sk-test",
            "http_proxy": "http://127.0.0.1:8080",
            "socks5_proxy": None,
            "enable_proxy": True,
            "max_concurrent_requests": 200,
            "is_active": True,
        })
        assert resp.status_code == 200
        data = resp.json()
        assert data["http_proxy"] == "http://127.0.0.1:8080"
        assert data["max_concurrent_requests"] == 200
        assert data["enable_proxy"] is True

    def test_create_provider_with_socks5_proxy(self, client, clean_tables):
        resp = client.post("/api/v1/providers/", json={
            "name": "SocksProvider",
            "base_url": "https://api.openai.com",
            "api_token": "sk-test",
            "http_proxy": None,
            "socks5_proxy": "127.0.0.1:1080",
            "enable_proxy": True,
            "max_concurrent_requests": 150,
            "is_active": True,
        })
        assert resp.status_code == 200
        data = resp.json()
        assert data["socks5_proxy"] == "127.0.0.1:1080"
        assert data["max_concurrent_requests"] == 150
        assert data["enable_proxy"] is True

    def test_update_provider_enable_proxy(self, client, clean_tables):
        # Create provider with proxy disabled
        create = client.post("/api/v1/providers/", json={
            "name": "ToggleProvider",
            "base_url": "https://api.openai.com",
            "api_token": "sk-test",
            "http_proxy": "http://127.0.0.1:8080",
            "enable_proxy": False,
            "max_concurrent_requests": 75,
            "is_active": True,
        })
        provider_id = create.json()["id"]

        # Toggle enable_proxy to True
        resp = client.put(f"/api/v1/providers/{provider_id}", json={
            "enable_proxy": True,
        })
        assert resp.status_code == 200
        assert resp.json()["enable_proxy"] is True
        assert resp.json()["http_proxy"] == "http://127.0.0.1:8080"

        # Toggle back to False
        resp = client.put(f"/api/v1/providers/{provider_id}", json={
            "enable_proxy": False,
        })
        assert resp.status_code == 200
        assert resp.json()["enable_proxy"] is False

    def test_list_providers(self, client, clean_tables):
        client.post("/api/v1/providers/", json={
            "name": "P1", "base_url": "https://a.com", "api_token": "t1",
        })
        client.post("/api/v1/providers/", json={
            "name": "P2", "base_url": "https://b.com", "api_token": "t2",
        })
        resp = client.get("/api/v1/providers/")
        assert resp.status_code == 200
        assert len(resp.json()) == 2

    def test_delete_provider(self, client, clean_tables):
        create = client.post("/api/v1/providers/", json={
            "name": "DeleteMe", "base_url": "https://a.com", "api_token": "t1",
        })
        provider_id = create.json()["id"]
        resp = client.delete(f"/api/v1/providers/{provider_id}")
        assert resp.status_code == 200
        resp = client.get(f"/api/v1/providers/{provider_id}")
        assert resp.status_code == 404


class TestModels:
    """Test model CRUD."""

    def test_create_model(self, client, clean_tables):
        provider = client.post("/api/v1/providers/", json={
            "name": "P1", "base_url": "https://a.com", "api_token": "t1",
        }).json()

        resp = client.post("/api/v1/models/", json={
            "provider_id": provider["id"],
            "model_id": "gpt-4",
            "is_active": True,
        })
        assert resp.status_code == 200
        data = resp.json()
        assert data["model_id"] == "gpt-4"
        assert data["provider_id"] == provider["id"]
        assert data["is_active"] is True

    def test_list_models(self, client, clean_tables):
        provider = client.post("/api/v1/providers/", json={
            "name": "P1", "base_url": "https://a.com", "api_token": "t1",
        }).json()
        client.post("/api/v1/models/", json={
            "provider_id": provider["id"], "model_id": "gpt-4", "is_active": True,
        })
        client.post("/api/v1/models/", json={
            "provider_id": provider["id"], "model_id": "gpt-3.5", "is_active": True,
        })

        resp = client.get("/api/v1/models/")
        assert resp.status_code == 200
        assert len(resp.json()) == 2

    def test_list_models_by_provider(self, client, clean_tables):
        p1 = client.post("/api/v1/providers/", json={
            "name": "P1", "base_url": "https://a.com", "api_token": "t1",
        }).json()
        p2 = client.post("/api/v1/providers/", json={
            "name": "P2", "base_url": "https://b.com", "api_token": "t2",
        }).json()
        client.post("/api/v1/models/", json={
            "provider_id": p1["id"], "model_id": "gpt-4", "is_active": True,
        })
        client.post("/api/v1/models/", json={
            "provider_id": p2["id"], "model_id": "gpt-3.5", "is_active": True,
        })

        resp = client.get(f"/api/v1/models/?provider_id={p1['id']}")
        assert resp.status_code == 200
        models = resp.json()
        assert len(models) == 1
        assert models[0]["model_id"] == "gpt-4"


class TestTokens:
    """Test token CRUD and permissions."""

    def test_create_token(self, client, clean_tables):
        resp = client.post("/api/v1/tokens/", json={
            "name": "TestToken",
            "max_input_tokens": 1000,
            "max_output_tokens": 500,
            "requests_per_minute": 60,
            "is_active": True,
            "model_permissions": [],
        })
        assert resp.status_code == 200
        data = resp.json()
        assert data["name"] == "TestToken"
        assert data["plain_key"].startswith("rpm_")
        assert data["max_input_tokens"] == 1000

    def test_create_token_with_model_permissions(self, client, clean_tables):
        provider = client.post("/api/v1/providers/", json={
            "name": "P1", "base_url": "https://a.com", "api_token": "t1",
        }).json()
        model = client.post("/api/v1/models/", json={
            "provider_id": provider["id"], "model_id": "gpt-4", "is_active": True,
        }).json()

        resp = client.post("/api/v1/tokens/", json={
            "name": "LimitedToken",
            "is_active": True,
            "model_permissions": [{"model_id": model["id"], "max_input_tokens": 1000, "max_output_tokens": 500}],
        })
        assert resp.status_code == 200
        data = resp.json()
        perm_model_ids = [p["model_id"] for p in data["model_permissions"]]
        assert model["id"] in perm_model_ids

    def test_delete_token(self, client, clean_tables):
        token = client.post("/api/v1/tokens/", json={
            "name": "DeleteMe", "is_active": True, "model_permissions": [],
        }).json()
        resp = client.delete(f"/api/v1/tokens/{token['id']}")
        assert resp.status_code == 200
        resp = client.get(f"/api/v1/tokens/{token['id']}")
        assert resp.status_code == 404


class TestConfig:
    """Test config endpoints."""

    def test_get_config(self, client, clean_tables):
        # Seed config value first (test DB doesn't run migrations)
        client.put("/api/v1/config/max_concurrent_requests", json={"value": "100"})
        resp = client.get("/api/v1/config/max_concurrent_requests")
        assert resp.status_code == 200
        assert resp.json()["key"] == "max_concurrent_requests"
        assert resp.json()["value"] == "100"

    def test_update_config(self, client, clean_tables):
        resp = client.put("/api/v1/config/max_concurrent_requests", json={
            "value": "200",
        })
        assert resp.status_code == 200
        assert resp.json()["value"] == "200"

        resp = client.get("/api/v1/config/max_concurrent_requests")
        assert resp.json()["value"] == "200"


class TestDashboard:
    """Test dashboard stats."""

    def test_dashboard_stats(self, client, clean_tables):
        resp = client.get("/api/v1/dashboard/stats")
        assert resp.status_code == 200
        data = resp.json()
        assert "total_requests" in data
        assert "active_providers" in data
        assert "active_tokens" in data
