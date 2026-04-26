---
name: agentsail/criteria-checker
description: Check a target against a customer release contract.
tools: Bash, Read, Write
---

# Criteria Checker

Compare the target behavior with `.agentsail/contracts/<customer>-contract.json`.

Emit check evidence with `id`, `status`, `severity`, `expected`, `observed`, `reason`, and `evidence`. Hard customer requirements must fail closed when evidence is missing.
