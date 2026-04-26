# empty-app — Loadtest Target Reference

Agent Sail 부하테스트 대상 서버 레퍼런스. LangGraph + Azure OpenAI + FastAPI + Prometheus.

`agentsail loadtest tui` 가 두드릴 두 endpoint 만 노출:

- `POST /chat` — `{"messages":[{"role":"user","content":"..."}]}`, 응답은 SSE (`data: <token>\n\n` … `data: [DONE]`)
- `GET /metrics` — Prometheus exposition (ttft / inter_token_latency / total_response / errors / requests / queue_depth / concurrent_*)

## 실행

```bash
cp .env.example .env   # AZURE_* 채우기
docker compose up -d --build
```

확인:

```bash
curl -sN -X POST http://localhost:8000/chat \
  -H 'content-type: application/json' \
  -d '{"messages":[{"role":"user","content":"say hi"}]}'

curl -s http://localhost:8000/metrics | grep ^ttft_seconds_count
```

## 부하테스트 (smoke / 30s × 4 users)

```bash
agentsail loadtest install --config agentsail.loadtest.smoke.yaml
agentsail loadtest tui     --config agentsail.loadtest.smoke.yaml --no-install
```

또는 `tools/release_board.py` (rich.Live, proposal §7 ASCII 그대로):

```bash
.agentsail/loadtests/.venv/bin/python tools/release_board.py \
  agentsail.loadtest.smoke.yaml acme-bank
```

## cmux Side-Pane

```bash
SURFACE=$(cmux --json new-split right | jq -r '.surface_ref')
cmux send --surface "$SURFACE" \
  ".agentsail/loadtests/.venv/bin/python tools/release_board.py agentsail.loadtest.smoke.yaml acme-bank"
cmux send-key --surface "$SURFACE" Return
```

## 정리

```bash
docker compose down
```

## 구조

```text
app/
  Dockerfile          # python:3.12-slim + requirements.txt
  requirements.txt    # langgraph, langchain-openai, fastapi, prometheus-client
  main.py             # FastAPI: /chat (SSE astream), /metrics, /healthz
docker-compose.yml    # 1g 메모리 limit, .env 로드
agentsail.loadtest.yaml          # 32 users / 12m (full)
agentsail.loadtest.smoke.yaml    # 4 users / 30s (smoke)
locust/agentsail/locustfile.py   # SSE-aware Locust user
tools/release_board.py           # rich.Live 4-panel TUI
```

## 메트릭 정의

`app/main.py` 가 노출하는 Prometheus 메트릭:

| Metric | Type | SLO |
|---|---|---|
| `ttft_seconds{route,mode}` | histogram | p95 < 1.5s |
| `inter_token_latency_seconds{route}` | histogram | p95 < 0.08s |
| `total_response_seconds{route,status}` | histogram | p95 < 10s |
| `llm_requests_total{route,status}` | counter | denominator |
| `llm_errors_total{route,kind}` | counter | < 1% |
| `request_queue_depth` | gauge | leading |
| `concurrent_llm_calls` | gauge | leading |
| `concurrent_sessions` | gauge | leading |
