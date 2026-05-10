import httpx
from typing import List, Dict, Any, Optional


def _normalize_socks5(addr: str) -> str:
    """Strip socks5:// prefix if present to avoid socks5://socks5://..."""
    addr = addr.strip()
    addr = addr.removeprefix("socks5://")
    return addr


def _normalize_http_proxy(addr: str) -> str:
    """Ensure http proxy has a scheme."""
    addr = addr.strip()
    if "://" not in addr:
        addr = "http://" + addr
    return addr


def fetch_models_from_provider(
    base_url: str,
    api_token: str,
    enable_proxy: bool = True,
    http_proxy: Optional[str] = None,
    socks5_proxy: Optional[str] = None,
) -> List[Dict[str, Any]]:
    # Handle base URLs that already end with /v1
    base = base_url.rstrip("/")
    if base.endswith("/v1"):
        url = base + "/models"
    else:
        url = base + "/v1/models"
    headers = {"Authorization": f"Bearer {api_token}"}

    mounts = None
    if enable_proxy:
        if http_proxy:
            proxy_url = _normalize_http_proxy(http_proxy)
            mounts = {
                "http://": httpx.HTTPTransport(proxy=proxy_url),
                "https://": httpx.HTTPTransport(proxy=proxy_url),
            }
        elif socks5_proxy:
            proxy_url = "socks5://" + _normalize_socks5(socks5_proxy)
            mounts = {
                "http://": httpx.HTTPTransport(proxy=proxy_url),
                "https://": httpx.HTTPTransport(proxy=proxy_url),
            }

    with httpx.Client(timeout=30.0, mounts=mounts) as client:
        resp = client.get(url, headers=headers)
        resp.raise_for_status()
        data = resp.json()
        return data.get("data", [])
