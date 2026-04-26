---
description: Collect customer context for Agent Sail.
argument-hint: "<customer>"
allowed-tools: Bash, Read, Write, Task
---

# /agentsail:collect

Use `agentsail-collect` for the requested customer. Write only `.agentsail/cache/<customer>/{slack,gmail,notion,github}.json`.

If live MCP data is unavailable, use concise fixture-style JSON and mark evidence confidence.
