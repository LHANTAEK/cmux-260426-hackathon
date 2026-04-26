# Agent Sail MVP Implementation Checklist

## Completed

- [x] Go CLI entrypoint: `cmd/agentsail/main.go`
- [x] Manual stdlib subcommand routing: `init`, `collect`, `compile`, `check`, `verdict`, `ci`, `report`, `loadtest`, `doctor`, `version`
- [x] Project init installer with embedded harness assets
- [x] Claude Code plugin and slash commands under `.claude/`
- [x] Codex command recipes under `.codex/commands/agentsail/`
- [x] Codex skill under `.codex/skills/agentsail/SKILL.md`
- [x] Local fixture collection into `.agentsail/cache/<customer>/`
- [x] Contract compilation into `.agentsail/contracts/<customer>-contract.json`
- [x] Mock target adapter for deterministic demos
- [x] HTTP target adapter and `examples/http_target/main.go`
- [x] Deterministic checks for citations, language, load SLO, 429 fallback, PII, CSV export, white label, tone
- [x] Verdict engine: `SHIP`, `HOLD`, `BLOCK`
- [x] Standalone HTML evidence report
- [x] cmux OSC9 alert support
- [x] YAML-driven Locust load-test template
- [x] Auto-install Locust/httpx with `uv` into `.agentsail/loadtests/.venv`
- [x] `agentsail loadtest tui` live terminal board for load-test demos
- [x] Docker Compose-style memory config: `limit: 1g`, `alert_at: 80%`
- [x] Demo verdicts: `finbank -> HOLD`, `retailco -> SHIP`, `acme-bank -> BLOCK`

## Verification

```bash
make test
make build
./bin/agentsail version
./bin/agentsail init /tmp/agentsail-project-test
cd /tmp/agentsail-project-test
agentsail ci --customer acme-bank --target mock:support_agent_v12 --report --cmux-alert --soft-exit
agentsail loadtest run --config agentsail.loadtest.yaml --dry-run
agentsail loadtest tui --config agentsail.loadtest.yaml --dry-run
```

## Remaining Product Work

- [ ] Real Slack/Notion/Gmail/GitHub MCP collectors beyond fixture ingest
- [ ] PR check-run integration
- [ ] Rich TUI release board
- [ ] Full Prometheus query evaluation for YAML SLO pass/fail after Locust run
- [ ] Release packaging script for GitHub binary artifacts
