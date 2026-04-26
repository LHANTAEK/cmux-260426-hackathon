# Agent Sail Collect

Collect customer launch context into `.agentsail/cache/<customer>/`.

## Inputs

- Customer slug, for example `finbank`, `retailco`, or `acme-bank`.

## Workflow

1. Create `.agentsail/cache/<customer>/`.
2. Run collector subagents in parallel when possible:
   - `agentsail/slack-collector`
   - `agentsail/notion-collector`
   - `agentsail/email-collector`
   - `agentsail/github-collector`
3. Add fixture or manual notes only when live MCP data is unavailable.
4. Write compact JSON files named by source: `slack.json`, `gmail.json`, `notion.json`, `github.json`.

## Output Shape

Each JSON file should contain `customer`, `source`, `collected_at`, and `items`. Preserve direct evidence snippets and links when available.
