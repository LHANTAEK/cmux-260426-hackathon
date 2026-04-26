# Agent Sail Runbook

Agent Sail is a Go CLI-first release gate with project-local Claude Code and Codex harness files.

## Install CLI

Web install from GitHub:

```bash
curl -fsSL https://raw.githubusercontent.com/LHANTAEK/cmux-260426-hackathon/mvp/scripts/install.sh | bash
```

The raw GitHub URL requires a public repository. Private repos return `404` for unauthenticated raw downloads.

Release binaries are published automatically by `.github/workflows/release.yml` when a `v*` tag is pushed:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Development install from this repo:

```bash
./scripts/install-local.sh
```

This installs `agentsail` into `~/.local/bin` by default. Override with:

```bash
AGENTSAIL_INSTALL_DIR=/tmp/bin ./scripts/install-local.sh
```

## Initialize A Project

Run once per project:

```bash
agentsail init
```

This installs:

- `.claude-plugin/` and `.claude/` commands, skills, agents, hooks, rules, settings
- `.codex/commands/agentsail/` command recipes
- `.codex/skills/agentsail/SKILL.md`
- `fixtures/agentsail/` demo customer context
- `agentsail.loadtest.yaml` and `locust/agentsail/`
- `.agentsail/{cache,contracts,runs,reports}` evidence directories
- an `AGENTS.md` Agent Sail block

## Claude Code Commands

```text
/agentsail
/agentsail:init
/agentsail:collect
/agentsail:compile
/agentsail:check
/agentsail:verdict
/agentsail:ci
/agentsail:report
/agentsail:loadtest
/agentsail:doctor
/agentsail:version
```

## Codex Integration

Codex uses project-local skills and recipes:

```text
.codex/skills/agentsail/SKILL.md
.codex/commands/agentsail/*.md
```

Codex should read those files and execute the terminal CLI directly from the project root.

## Demo Commands

```bash
agentsail ci --customer finbank --target mock:support_agent_v12 --report --soft-exit
agentsail ci --customer retailco --target mock:support_agent_v12 --report
agentsail ci --customer acme-bank --target mock:support_agent_v12 --report --cmux-alert --soft-exit
```

Expected verdicts:

```text
finbank   -> HOLD
retailco  -> SHIP
acme-bank -> BLOCK
```

Artifacts:

```text
.agentsail/cache/<customer>/*
.agentsail/contracts/<customer>-contract.json
.agentsail/runs/<customer>-run-NNN.json
.agentsail/reports/<customer>-run-NNN.html
```

## Load Testing

Generate or refresh the YAML template and Locust files:

```bash
agentsail loadtest init --config agentsail.loadtest.yaml
```

Explain the metric model:

```bash
agentsail loadtest explain
```

Install the project-local runtime explicitly, or let `run` install it automatically:

```bash
agentsail loadtest install --config agentsail.loadtest.yaml
```

Run or preview Locust:

```bash
agentsail loadtest run --config agentsail.loadtest.yaml --dry-run
agentsail loadtest tui --config agentsail.loadtest.yaml --dry-run
agentsail loadtest tui --config agentsail.loadtest.yaml
agentsail loadtest run --config agentsail.loadtest.yaml
```

Use `tui` for live load-test usability checks. It runs Locust and renders a terminal board with target, user profile, SLOs, memory limit/alert in GB, artifact paths, and recent Locust output.

`run` and `tui` auto-install `locust` and `httpx` with `uv` into `.agentsail/loadtests/.venv` when missing.

## Load-Test YAML

Users configure memory like Docker Compose, not as raw bytes:

```yaml
resources:
  memory:
    limit: 1g
    alert_at: 80%
```

Metric defaults are based on `/Users/pangpang/devops/llm-apps-monitoring-0424/main/docs/02-metrics.md`:

| Metric | Meaning | Default SLO |
|---|---|---|
| `ttft_seconds` | Time to first token | p95 `< 1.5s` |
| `inter_token_latency_seconds` | SSE inter-token gap | p95 `< 0.08s` |
| `total_response_seconds` | Request to last token wall clock | p95 `< 10s` |
| `llm_errors_total / llm_requests_total` | Error rate | `< 1%` |
| `container_memory_working_set_bytes` | Container memory working set | `< resources.memory.alert_at` of `resources.memory.limit` |
| `request_queue_depth` | Queue pressure leading signal | candidate threshold `4` |
| `concurrent_llm_calls` | LLM semaphore saturation | candidate ratio `0.75` |
| `concurrent_sessions` | Active SSE sessions | observed leading signal |
