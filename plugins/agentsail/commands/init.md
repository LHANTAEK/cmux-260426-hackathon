---
description: Initialize Agent Sail harness files in the current project.
allowed-tools: [Read, Glob, Grep, Bash]
---

# /agentsail:init

Install Agent Sail into the current project:

```bash
agentsail init .
agentsail loadtest init --config agentsail.loadtest.yaml
```

After init, Codex should use the terminal CLI:

```bash
agentsail ci --customer acme-bank --target mock:support_agent_v12 --report --cmux-alert
```

