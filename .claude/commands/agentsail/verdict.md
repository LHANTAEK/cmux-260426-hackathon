---
description: Print the latest SHIP/HOLD/BLOCK verdict for a customer.
argument-hint: "--customer <customer>"
allowed-tools: Bash, Read, Write, Edit
---

# /agentsail:verdict

Convert check evidence into `SHIP`, `HOLD`, or `BLOCK`.

## Workflow

1. Read the latest run JSON for the customer.
2. Use `agentsail/verdict-engine`.
3. Apply `.claude/rules/agentsail-release-gate.md`.
4. Prefer the Go CLI:
   - `bin/agentsail verdict --customer <customer>`
   - fallback: `go run ./cmd/agentsail verdict --customer <customer>`

```bash
agentsail verdict --customer acme-bank
```

## Decision

- `BLOCK` for missing required capability, exposed forbidden state, or severe tone drift.
- `HOLD` for incomplete evidence or non-blocking mismatch.
- `SHIP` only when required checks pass with evidence.
