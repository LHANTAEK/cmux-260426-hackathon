---
description: Render an Agent Sail report from a run JSON.
argument-hint: "<run-json> [--open]"
allowed-tools: Bash, Read, Write
---

# /agentsail:report

Render HTML evidence for a completed run.

## Workflow

1. Read the run JSON path from the request.
2. Use `agentsail/report-renderer`.
3. Prefer the Go CLI:
   - `bin/agentsail report <run-json> [--open]`
   - fallback: `go run ./cmd/agentsail report <run-json> [--open]`
4. Confirm `.agentsail/reports/<run-id>.html` exists.

```bash
bin/agentsail report <run-json> --open
```

## Report Contents

Include verdict, failed reasons, customer contract summary, and evidence snippets. Keep it standalone HTML.
