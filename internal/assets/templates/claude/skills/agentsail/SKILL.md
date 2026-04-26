# Agent Sail Dispatcher

Use this skill for any `/agentsail` request.

## Dispatch

- `collect <customer>` -> use `agentsail-collect`.
- `compile <customer>` -> use `agentsail-compile`.
- `check <customer> --target <url>` -> use `agentsail-check`.
- `verdict <customer>` -> use `agentsail-verdict`.
- `report <run-json> [--open]` -> use `agentsail-report`.
- `ci --customer <customer> --target <url> [--report] [--open]` -> use `agentsail-ci`.
- `loadtest tui [--config agentsail.loadtest.yaml]` -> run the live Locust terminal board.

## Rules

- Keep state under `.agentsail/{cache,contracts,runs,reports}`.
- Prefer `bin/agentsail`; fall back to `go run ./cmd/agentsail`.
- If cache is missing, tell the user to run `/agentsail collect <customer>` first.
- Keep verdict wording exactly `SHIP`, `HOLD`, or `BLOCK`.
- For live load-test usability checks, prefer `agentsail loadtest tui --config agentsail.loadtest.yaml`.
