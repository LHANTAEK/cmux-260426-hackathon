"""Agent Sail SSE-aware Locust load test.

Metric/SLO references:
- ttft_seconds: p95 < 1.5s
- inter_token_latency_seconds: p95 < 80ms
- total_response_seconds: p95 < 10s
- llm_errors_total / llm_requests_total: < 1%
- request_queue_depth, concurrent_llm_calls, concurrent_sessions: leading signals

The real production metrics should come from the app's /metrics endpoint and
Prometheus/VictoriaMetrics. Locust generates pressure and writes CSV/HTML
artifacts under .agentsail/loadtests.
"""

from __future__ import annotations

import os
import random
import time

import httpx
from locust import HttpUser, between, events, task

SHORT_PROMPTS = [
    "What does TTFT mean?",
    "Define latency in one sentence.",
    "Name one metric for queue pressure.",
    "Give me a one-word greeting.",
]

LONG_PROMPTS = [
    "Explain in three sentences why p95 latency matters more than mean for a streaming chat application.",
    "Describe the difference between time-to-first-token and inter-token latency in three sentences.",
]


class AgentSailChatUser(HttpUser):
    host = os.environ.get("TARGET_HOST", "http://localhost:8000")
    wait_time = between(0.5, 2.0)

    @task(8)
    def short_prompt(self) -> None:
        self._chat(random.choice(SHORT_PROMPTS), name="chat:short")

    @task(2)
    def long_prompt(self) -> None:
        self._chat(random.choice(LONG_PROMPTS), name="chat:long")

    def _chat(self, prompt: str, *, name: str) -> None:
        chat_path = os.environ.get("CHAT_PATH", "/chat")
        payload = {"messages": [{"role": "user", "content": prompt}]}
        start = time.perf_counter()
        first_token_at: float | None = None
        token_count = 0
        exception: Exception | None = None
        status = 0

        try:
            with httpx.stream(
                "POST",
                f"{self.host}{chat_path}",
                json=payload,
                timeout=httpx.Timeout(60.0, connect=5.0),
            ) as resp:
                status = resp.status_code
                if status >= 400:
                    resp.read()
                    raise httpx.HTTPStatusError(
                        f"upstream {status}", request=resp.request, response=resp
                    )
                for line in resp.iter_lines():
                    if not line or not line.startswith("data:"):
                        continue
                    if first_token_at is None:
                        first_token_at = time.perf_counter()
                    token_count += 1
        except Exception as exc:
            exception = exc

        elapsed = time.perf_counter() - start
        events.request.fire(
            request_type="SSE",
            name=name,
            response_time=elapsed * 1000,
            response_length=token_count,
            exception=exception,
            context={
                "ttft_ms": (first_token_at - start) * 1000 if first_token_at else None,
                "status": status,
            },
        )

