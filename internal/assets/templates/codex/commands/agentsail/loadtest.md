# agentsail:loadtest

YAML-driven Locust load test workflow:

```bash
agentsail loadtest init --config agentsail.loadtest.yaml
agentsail loadtest explain
agentsail loadtest doctor --config agentsail.loadtest.yaml
agentsail loadtest install --config agentsail.loadtest.yaml
agentsail loadtest run --config agentsail.loadtest.yaml --dry-run
agentsail loadtest tui --config agentsail.loadtest.yaml --dry-run
agentsail loadtest tui --config agentsail.loadtest.yaml
agentsail loadtest run --config agentsail.loadtest.yaml
```

Use `tui` when the user wants to see the load test while it runs. It shows the target, user profile, SLOs, memory limit/alert in GB, artifact paths, and recent Locust output in the terminal.

`run` and `tui` auto-installs Locust and httpx with `uv` into `.agentsail/loadtests/.venv` when missing. Use `--no-install` only when the CI image already owns the runtime.

The YAML template documents the LLM app monitoring metrics and SLOs:

- `ttft_seconds` p95 < 1.5s
- `inter_token_latency_seconds` p95 < 80ms
- `total_response_seconds` p95 < 10s
- `llm_errors_total / llm_requests_total` < 1%
- `request_queue_depth`, `concurrent_llm_calls`, `concurrent_sessions` as leading signals
- Use Docker Compose-style memory values in YAML: `resources.memory.limit: 1g` and `alert_at: 80%`. Agent Sail converts internally.
