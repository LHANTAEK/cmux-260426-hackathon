---
description: Generate or run Agent Sail Locust load tests from YAML.
argument-hint: "<init|run|explain> [--config agentsail.loadtest.yaml]"
allowed-tools: Bash, Read, Write, Edit
---

# /agentsail:loadtest

Use YAML-driven Locust load tests:

```bash
agentsail loadtest init --config agentsail.loadtest.yaml
agentsail loadtest explain
agentsail loadtest doctor --config agentsail.loadtest.yaml
agentsail loadtest install --config agentsail.loadtest.yaml
agentsail loadtest run --config agentsail.loadtest.yaml --dry-run
agentsail loadtest run --config agentsail.loadtest.yaml
```

`run` auto-installs Locust and httpx into `.agentsail/loadtests/.venv` when missing. Use `--no-install` when a CI image must provide Locust itself.

Metrics tracked by the template come from `llm-apps-monitoring-0424`: `ttft_seconds`, `inter_token_latency_seconds`, `total_response_seconds`, `llm_requests_total`, `llm_errors_total`, `request_queue_depth`, `concurrent_llm_calls`, `concurrent_sessions`, and `container_memory_working_set_bytes`.

Use Docker Compose-style memory values in YAML: `resources.memory.limit: 1g` and `alert_at: 80%`. Agent Sail converts internally.
