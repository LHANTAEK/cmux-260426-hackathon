---
description: Dispatch Agent Sail release-gate workflows.
argument-hint: "<collect|compile|check|verdict|report|ci|doctor|version> [args]"
allowed-tools: Bash, Read, Write, Edit, Task
---

# /agentsail

Dispatch Agent Sail. Keep all generated evidence under `.agentsail/`.

Usage:

- `/agentsail collect <customer>`: fan out collector agents and write `.agentsail/cache/<customer>/*.json`.
- `/agentsail compile <customer>`: compile cached context into `.agentsail/contracts/<customer>-contract.json`.
- `/agentsail check <customer> --target <url>`: run deterministic gate checks.
- `/agentsail verdict <customer>`: produce `SHIP`, `HOLD`, or `BLOCK`.
- `/agentsail report <run-json> [--open]`: render HTML evidence.
- `/agentsail ci --customer <customer> --target <url> [--report] [--open]`: run compile, check, verdict, report sequentially.

When a Go CLI exists, prefer `bin/agentsail`; otherwise use `go run ./cmd/agentsail`. For collection, use the `agentsail-collect` skill and the collector subagents before invoking CI.
