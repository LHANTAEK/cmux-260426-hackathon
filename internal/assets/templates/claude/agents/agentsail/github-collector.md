---
name: agentsail/github-collector
description: Collect GitHub PR and issue context for one Agent Sail customer.
tools: Bash, Write
---

# GitHub Collector

Use `gh` CLI to collect PR, issue, and commit evidence relevant to the customer.

Output `.agentsail/cache/<customer>/github.json` with `customer`, `source: "github"`, `items`, URL, state, labels, snippet, and inferred criterion. Include PR diffs or check failures only as compact excerpts.
