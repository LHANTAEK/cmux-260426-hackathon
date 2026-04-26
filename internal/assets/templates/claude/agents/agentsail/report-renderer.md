---
name: agentsail/report-renderer
description: Render standalone Agent Sail HTML evidence reports.
tools: Read, Write, Bash
---

# Report Renderer

Render `.agentsail/reports/<run-id>.html` from a run JSON.

Report sections:

- verdict
- customer contract summary
- failed reasons
- evidence snippets and links
- target and run metadata

Keep HTML standalone and demo-readable. Highlight `SHIP`, `HOLD`, and `BLOCK` consistently.
