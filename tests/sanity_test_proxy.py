#!/usr/bin/env python3
"""
Comprehensive sanity and concurrency test for the OpenAI proxy server.

Tests:
  - /health endpoint
  - /v1/models list and individual model fetch
  - Single chat completion
  - Concurrent load: 1, 5, 10, 20, 50, 100, 128 concurrent requests
  - Varying prompt lengths: short, medium, long
  - Reports latency percentiles, success rates, and error breakdowns

Usage:
    python tests/sanity_test_proxy.py
"""

import concurrent.futures
import requests
import sys
import time
from collections import Counter

# ── Configuration ──────────────────────────────────────────────────────────
BASE_URL = "http://localhost:8080"
TOKEN = "rpm_TsW1s7faYT9-pXSuBYt4UAo--x1ApaTmBvo_hlgUILE"
MODEL = "xerxes-8.19b"
TIMEOUT = 60  # seconds per request

HEADERS = {
    "Authorization": f"Bearer {TOKEN}",
    "Content-Type": "application/json",
}


def make_chat_request(prompt_len: int, max_tokens: int = 10) -> dict:
    """Send a single chat-completion request. Returns result dict."""
    prompt = "word " * prompt_len
    body = {
        "model": MODEL,
        "messages": [{"role": "user", "content": prompt}],
        "max_tokens": max_tokens,
    }
    t0 = time.time()
    try:
        resp = requests.post(
            f"{BASE_URL}/v1/chat/completions",
            headers=HEADERS,
            json=body,
            timeout=TIMEOUT,
        )
        latency = time.time() - t0
        if resp.status_code == 200:
            data = resp.json()
            usage = data.get("usage", {})
            return {
                "ok": True,
                "status": resp.status_code,
                "latency": latency,
                "prompt_len": prompt_len,
                "prompt_tokens": usage.get("prompt_tokens", 0),
                "completion_tokens": usage.get("completion_tokens", 0),
                "total_tokens": usage.get("total_tokens", 0),
                "error": None,
            }
        else:
            return {
                "ok": False,
                "status": resp.status_code,
                "latency": latency,
                "prompt_len": prompt_len,
                "prompt_tokens": 0,
                "completion_tokens": 0,
                "total_tokens": 0,
                "error": resp.text[:120],
            }
    except Exception as exc:
        latency = time.time() - t0
        return {
            "ok": False,
            "status": -1,
            "latency": latency,
            "prompt_len": prompt_len,
            "prompt_tokens": 0,
            "completion_tokens": 0,
            "total_tokens": 0,
            "error": str(exc)[:120],
        }


def run_concurrent_batch(concurrency: int, prompt_len: int, max_tokens: int = 10) -> list:
    """Run N concurrent chat requests with identical parameters."""
    with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as ex:
        futures = [
            ex.submit(make_chat_request, prompt_len, max_tokens)
            for _ in range(concurrency)
        ]
        return [f.result() for f in futures]


def print_batch_summary(label: str, results: list, wall_time: float):
    """Pretty-print statistics for a batch."""
    ok = [r for r in results if r["ok"]]
    failed = [r for r in results if not r["ok"]]
    latencies = [r["latency"] for r in ok]

    print(f"\n  {label}")
    print(f"    Wall time   : {wall_time:.2f}s")
    print(f"    Success     : {len(ok)}/{len(results)} ({len(ok)*100//len(results)}%)")

    if latencies:
        latencies.sort()
        p50 = latencies[len(latencies) // 2]
        p95 = latencies[int(len(latencies) * 0.95)] if len(latencies) > 1 else latencies[0]
        avg = sum(latencies) / len(latencies)
        total_tokens = sum(r["total_tokens"] for r in ok)
        print(f"    Latency     : avg={avg:.2f}s  p50={p50:.2f}s  p95={p95:.2f}s")
        print(f"    Total tokens: {total_tokens}")

    if failed:
        err_counts = Counter(r["error"] for r in failed)
        print(f"    Errors:")
        for err, cnt in err_counts.most_common(3):
            print(f"      [{cnt}x] {err}")


def main():
    print("=" * 60)
    print("OpenAI Proxy Server – Comprehensive Sanity Test")
    print("=" * 60)

    # ── 1. Health check ──────────────────────────────────────────────────
    print("\n[1] Health check")
    try:
        r = requests.get(f"{BASE_URL}/health", timeout=5)
        assert r.status_code == 200 and r.json().get("status") == "ok"
        print("    OK – server is healthy")
    except Exception as exc:
        print(f"    FAIL – {exc}")
        sys.exit(1)

    # ── 2. Models list ───────────────────────────────────────────────────
    print("\n[2] /v1/models list")
    try:
        r = requests.get(f"{BASE_URL}/v1/models", headers=HEADERS, timeout=15)
        assert r.status_code == 200
        models = r.json().get("data", [])
        print(f"    OK – {len(models)} model(s) visible to token")
        for m in models:
            print(f"        • {m.get('id')} ({m.get('owned_by', 'unknown')})")
    except Exception as exc:
        print(f"    FAIL – {exc}")

    # ── 3. Single chat completion ────────────────────────────────────────
    print("\n[3] Single chat completion")
    r = make_chat_request(prompt_len=5, max_tokens=10)
    if r["ok"]:
        print(f"    OK – latency={r['latency']:.2f}s tokens={r['total_tokens']}")
    else:
        print(f"    FAIL – status={r['status']} error={r['error']}")
        sys.exit(1)

    # ── 4. Varying prompt lengths (5 concurrent each) ────────────────────
    print("\n[4] Varying prompt lengths @ 5 concurrent")
    for plen, label in [(5, "short"), (20, "medium"), (100, "long"), (200, "very long")]:
        t0 = time.time()
        results = run_concurrent_batch(5, plen)
        print_batch_summary(f"prompt_len={plen:3d} ({label:9s})", results, time.time() - t0)

    # ── 5. Concurrency ramp-up ───────────────────────────────────────────
    print("\n[5] Concurrency ramp-up")
    levels = [1, 5, 10, 20, 50, 100, 128]
    for conc in levels:
        t0 = time.time()
        results = run_concurrent_batch(conc, prompt_len=10)
        wall = time.time() - t0
        ok_count = sum(1 for r in results if r["ok"])
        print_batch_summary(f"{conc:3d} concurrent", results, wall)
        if ok_count == 0 and conc > 1:
            print(f"    ! All requests failed at {conc} concurrent – stopping ramp-up.")
            break

    print("\n" + "=" * 60)
    print("Sanity test complete.")
    print("=" * 60)


if __name__ == "__main__":
    main()
