---
description: Install Agent Sail into this project.
argument-hint: "[path]"
allowed-tools: Bash, Read, Write, Edit
---

# /agentsail:init

Run project-local Agent Sail initialization:

```bash
agentsail init ${ARGUMENTS:-.}
agentsail loadtest init --config agentsail.loadtest.yaml
```

This installs `.claude/`, `.claude-plugin/`, `.codex/`, demo fixtures, Locust templates, `agentsail.loadtest.yaml`, and `.agentsail/` evidence directories.

