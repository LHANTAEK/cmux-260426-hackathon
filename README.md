# Agent Sail

Customer-aware release gate for AI agent demos.

## Install

Web install from GitHub:

```bash
curl -fsSL https://raw.githubusercontent.com/LHANTAEK/cmux-260426-hackathon/mvp/scripts/install.sh | bash
```

That URL requires the GitHub repository to be public. For a private repository, use an authenticated download or make the repository public before sharing the one-line installer.

Release binaries are published automatically when a `v*` tag is pushed:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Local development install:

```bash
./scripts/install-local.sh
```

This installs `agentsail` to `~/.local/bin` by default. Override with:

```bash
AGENTSAIL_INSTALL_DIR=/tmp/bin ./scripts/install-local.sh
```

## Project Init

Run once per project:

```bash
agentsail init
```

This installs:

- `.claude-plugin/` and `.claude/` for Claude Code commands, skills, agents, hooks, and rules
- `.codex/commands/agentsail/` command recipes for Codex
- `fixtures/agentsail/` demo context for deterministic local runs
- `agentsail.loadtest.yaml` and `locust/agentsail/` for YAML-driven Locust load tests
- `.agentsail/` evidence directories
- an `AGENTS.md` Agent Sail block for Codex and other terminal agents

## Demo

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

Claude Code commands:

- `/agentsail:init`
- `/agentsail:collect`
- `/agentsail:compile`
- `/agentsail:check`
- `/agentsail:verdict`
- `/agentsail:ci`
- `/agentsail:report`
- `/agentsail:loadtest`
- `/agentsail:doctor`
- `/agentsail:version`

Codex command recipes are installed under `.codex/commands/agentsail/`. Codex should read those recipes and execute the `agentsail` terminal CLI directly.

## Load Test

Generate the YAML template and Locust files:

```bash
agentsail loadtest init --config agentsail.loadtest.yaml
```

Explain the metric model:

```bash
agentsail loadtest explain
```

Install the project-local Locust runtime explicitly, or let `run` install it automatically:

```bash
agentsail loadtest doctor --config agentsail.loadtest.yaml
agentsail loadtest install --config agentsail.loadtest.yaml
```

Preview or run Locust. `run` and `tui` create `.agentsail/loadtests/.venv` with `uv` and install `locust` + `httpx` when missing:

```bash
agentsail loadtest run --config agentsail.loadtest.yaml --dry-run
agentsail loadtest tui --config agentsail.loadtest.yaml --dry-run
agentsail loadtest tui --config agentsail.loadtest.yaml
agentsail loadtest run --config agentsail.loadtest.yaml
```

Use `agentsail loadtest tui` for live terminal usability tests. It runs Locust and shows the target, load profile, SLOs, memory alert converted to GB, artifact paths, and recent Locust output.

The YAML template documents metrics from `llm-apps-monitoring-0424`:

- `ttft_seconds`: p95 time to first token, SLO `< 1.5s`
- `inter_token_latency_seconds`: p95 inter-token latency, SLO `< 0.08s`
- `total_response_seconds`: p95 total response time, SLO `< 10s`
- `llm_errors_total / llm_requests_total`: error rate, SLO `< 1%`
- `request_queue_depth`, `concurrent_llm_calls`, `concurrent_sessions`: leading saturation signals
- `container_memory_working_set_bytes`: memory working set; configure `resources.memory.limit: 1g` and `alert_at: 80%`
