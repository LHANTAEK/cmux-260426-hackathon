---
name: agentsail
description: "Use when running Agent Sail in Codex: project init, customer release gates, evidence reports, and YAML-driven Locust load tests."
---

# Agent Sail

Use the terminal CLI from the project root. Keep all generated evidence under `.agentsail/`.

## Commands

```bash
agentsail init .
agentsail collect acme-bank
agentsail compile --customer acme-bank
agentsail check --customer acme-bank --target mock:support_agent_v12
agentsail verdict --customer acme-bank
agentsail ci --customer acme-bank --target mock:support_agent_v12 --report --cmux-alert
agentsail report .agentsail/runs/acme-bank-run-001.json --open
agentsail loadtest init --config agentsail.loadtest.yaml
agentsail loadtest tui --config agentsail.loadtest.yaml --dry-run
agentsail loadtest tui --config agentsail.loadtest.yaml
agentsail loadtest run --config agentsail.loadtest.yaml
```

Use `agentsail loadtest tui` for live load-test demos and usability testing. It is still a terminal CLI command, but it renders a live board with target, concurrency, SLOs, memory threshold converted to GB, artifact paths, and recent Locust output.

## Load Test YAML

Memory values are Docker Compose-style human values:

```yaml
resources:
  memory:
    limit: 1g
    alert_at: 80%
```

Do not ask users to write raw byte values. Agent Sail converts values internally.
