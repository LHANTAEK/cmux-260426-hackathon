---
description: Run Agent Sail CI release gate.
argument-hint: "--customer <customer> --target <url> [--report] [--open]"
allowed-tools: Bash, Read, Write
---

# /agentsail:ci

Use `agentsail-ci`. Prefer:

```bash
bin/agentsail ci --customer <customer> --target <url> --report --open
```

Fall back to `go run ./cmd/agentsail ci ...`. Stop if `.agentsail/cache/<customer>/` is empty.
