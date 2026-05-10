from fastapi import APIRouter
from fastapi.responses import PlainTextResponse
import httpx
from app.config import settings

router = APIRouter()

PROXY_BASE = settings.proxy_base_url.rstrip("/")


@router.get("/status", tags=["proxy"])
async def proxy_status():
    """Forward Go proxy /api/status JSON."""
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{PROXY_BASE}/api/status", timeout=10.0)
    return resp.json()


@router.get("/logs", tags=["proxy"])
async def proxy_logs():
    """Forward Go proxy /api/logs JSON."""
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{PROXY_BASE}/api/logs", timeout=10.0)
    return resp.json()


@router.delete("/logs", tags=["proxy"], status_code=204)
async def clear_proxy_logs():
    """Forward Go proxy /api/logs DELETE."""
    async with httpx.AsyncClient() as client:
        await client.delete(f"{PROXY_BASE}/api/logs", timeout=10.0)


@router.get("/metrics", response_class=PlainTextResponse, tags=["proxy"])
async def proxy_metrics():
    """Forward Go proxy /metrics Prometheus text."""
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{PROXY_BASE}/metrics", timeout=10.0)
    return PlainTextResponse(content=resp.text, media_type="text/plain")
