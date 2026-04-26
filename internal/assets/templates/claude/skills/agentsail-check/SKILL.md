# Agent Sail Check

Run deterministic release checks for one customer and target.

## Workflow

1. Confirm `.agentsail/contracts/<customer>-contract.json` exists.
2. Use `agentsail/criteria-checker` for hard rules.
3. Use `agentsail/chaos-prober` for timeout, 429, and empty-retrieval probes.
4. Prefer the Go CLI:
   - `bin/agentsail check --customer <customer> --target <url>`
   - fallback: `go run ./cmd/agentsail check --customer <customer> --target <url>`

## Evidence

Record pass/fail checks with reason, source evidence, and observed target response.
