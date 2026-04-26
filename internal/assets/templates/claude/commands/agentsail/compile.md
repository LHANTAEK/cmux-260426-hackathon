---
description: Compile cached customer context into a release contract.
argument-hint: "--customer <customer>"
allowed-tools: Bash, Read, Write, Edit
---

# /agentsail:compile

Compile cached customer context into a release contract.

## Workflow

1. Read `.agentsail/cache/<customer>/*.json`.
2. Use the `agentsail/contract-compiler` subagent for synthesis.
3. Prefer the Go CLI when available:
   - `bin/agentsail compile --customer <customer>`
   - fallback: `go run ./cmd/agentsail compile --customer <customer>`
4. Ensure `.agentsail/contracts/<customer>-contract.json` exists.

```bash
agentsail compile --customer acme-bank
```

## Contract Fields

Keep the contract small: `customer`, `required_capabilities`, `forbidden_exposures`, `tone`, `slo`, `evidence`. Keep generated evidence under `.agentsail/`.
