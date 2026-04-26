# Agent Sail Release Gate Rules

Agent Sail turns customer context into an evidence-backed release verdict.

- `SHIP`: all hard rules pass, no customer-specific blockers, report generated.
- `HOLD`: non-blocking criteria fail or evidence is incomplete; human review required.
- `BLOCK`: hard customer rule fails, unsafe branding/tone is exposed, or required feature is missing.

The MVP story must preserve:

- Same agent. Different customer. Different launch gate.
- `finbank` -> `HOLD`
- `retailco` -> `SHIP`
- `acme-bank` -> `BLOCK` with `missing CSV export`, `beta badge exposed`, and `tone drift`.

Do not claim production SaaS connectors. Collect through Claude Code MCP, `gh`, or local fixtures, then write normalized cache JSON.
