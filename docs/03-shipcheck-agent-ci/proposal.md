# 03. Agent Sail — Customer-specific Agent CI

## 1. Problem

**AI agents can build fast, but teams still do not know whether the result is successful for a specific customer.**

In small AI teams, customer success criteria are scattered across meeting notes, PRDs, Slack threads, demo feedback, and tickets. A coding agent can produce a working demo, but the team often ships without checking it against the criteria that actually matter for that customer.

Examples:

- FinBank requires citations in every answer, no PII exposure, and 50 concurrent internal users during onboarding.
- RetailCo cares more about Korean refund-policy answers and only expects 20 concurrent users.
- MedOps requires escalation for medical/legal questions and approval before external messages.

The same agent can be shippable for one customer and a launch blocker for another.

## 2. Product Definition

**Agent Sail turns customer-specific success criteria into executable release gates for AI agents.**

Short version:

> Agent Sail checks whether an agent is ready to ship for this customer.

Positioning:

| Tool category | What it answers |
|---|---|
| Traditional CI | Does the code pass tests? |
| Observability | What happened after deployment? |
| LLM gateway | How should model requests be routed? |
| **Agent Sail** | **Does this agent satisfy this customer's launch criteria?** |

Key message:

> Code has CI before deployment. AI agents need customer-specific CI before launch.

## 3. Core Workflow

```
Customer context
  -> Customer success criteria
  -> Release gates
  -> Agent harness checks
  -> SHIP / HOLD / BLOCK verdict
```

Agent Sail is not a general QA dashboard and not an LLM gateway. It is the release gate between agent-built output and customer-specific acceptance.

## 4. Interface Decision

### 4.1 Primary interface: CLI/TUI

The core product should run in the terminal.

Reasons:

- The hackathon is centered on terminal workflows and coding agents.
- A release gate needs stdout, exit codes, and artifacts.
- Agent harness integration is simpler from a CLI.
- Live load/SLO probes can be shown directly in a TUI.
- It fits CI/CD and pre-merge workflows.

Primary command:

```bash
agentsail ci --customer finbank --target http://localhost:8000/chat
```

Expected behavior:

- Load customer criteria.
- Run agent harness checks.
- Run SLO/load probe.
- Run fallback/chaos-lite checks.
- Print live TUI status.
- Write structured result artifact.
- Exit non-zero on `HOLD` or `BLOCK`.

### 4.2 Secondary interface: static web report

Web should not be the main product for the hackathon MVP. It should be a generated report.

```bash
agentsail ci --customer finbank --target http://localhost:8000/chat --report
```

Generated artifact:

```text
.agentsail/reports/finbank-run-001.html
```

Purpose:

- Share evidence with non-terminal viewers.
- Show load curve, failed criteria, samples, and fix suggestions.
- Support the demo without turning the product into a dashboard.

## 5. TUI MVP

The TUI should be a release board, not a full whiteboard.

```
┌ Customers ───────┐ ┌ Criteria / Evidence ─────────────────────┐ ┌ Verdict ───────┐
│ FinBank          │ │ citations_required                        │ │ HOLD            │
│ RetailCo         │ │ source: meeting.md:12                     │ │                 │
│ MedOps           │ │                                           │ │ Failed:         │
│                  │ │ expected_concurrency: 50                  │ │ 1. citations    │
│ Target           │ │ p95_latency_ms: 5000                      │ │ 2. 31/50 users  │
│ langgraph:app    │ │ fallback_on_429: short_answer             │ │ 3. no fallback  │
└──────────────────┘ └───────────────────────────────────────────┘ └─────────────────┘

┌ Live Load Probe ───────────────────────────────────────────────────────────────┐
│ Phase: RAMPING     Users: 37 / 50     RPS: 12.4                               │
│ p50: 1.2s          p95: 5.8s          error: 0.4%                              │
│                                                                                │
│ Load      ▁▂▃▄▅▆▇                                                              │
│ p95       ▁▁▂▂▄▆█                                                              │
└────────────────────────────────────────────────────────────────────────────────┘
```

Minimum controls:

- `r`: run checks
- `c`: view contract
- `e`: view evidence
- `q`: quit

For hackathon reliability, this can be implemented with `rich.Live` first. A fuller Textual UI can be a stretch goal.

## 6. Customer Criteria Model

The MVP should support explicit customer criteria first. Automatic extraction from messy context is useful, but it is not required for the core demo.

Example:

```yaml
customer: finbank
target:
  type: http
  endpoint: http://localhost:8000/chat
  method: POST
  prompt_field: message
  answer_field: answer
  citations_field: citations

criteria:
  quality:
    citations_required: true

  reliability:
    expected_concurrency: 50
    max_p95_latency_ms: 5000
    max_error_rate: 0.01
    fallback_on_429: short_answer

  safety:
    external_email_requires_approval: true
```

LangGraph target example for local app checks:

```yaml
customer: finbank
target:
  type: langgraph
  module: examples.support_graph
  app: app
  input_key: message
  answer_field: answer
  citations_field: citations

criteria:
  reliability:
    expected_concurrency: 50
    max_p95_latency_ms: 5000
    max_error_rate: 0.01
    fallback_on_429: short_answer
```

Two demo customers:

- `FinBank`: strict citations, 50 concurrent users, fallback required.
- `RetailCo`: Korean refund answers, 20 concurrent users, fallback optional.

The same agent should produce different verdicts depending on the customer.

## 7. MVP Checks

### 7.1 Criteria check

Validate deterministic response requirements.

Examples:

- Response must include `citations`.
- Response language must be Korean.
- External email action must include `approval_required`.

### 7.2 Load/SLO probe

Run a short launch-readiness load probe, not a full production benchmark.

MVP approach:

- Use `asyncio` as the load driver.
- Use `httpx` for HTTP targets.
- Use the LangGraph adapter for local graph targets.
- Ramp users in short steps.
- Track p50, p95, error rate, completed requests.
- Compare against customer SLO.

Example result:

```text
[FAIL] load: reached only 31/50 users before p95 exceeded 5000ms
```

Language to use:

> This is a launch-readiness probe, not a production capacity certification.

### 7.3 Chaos-lite check

Inject one or two predictable failures through the harness adapter.

MVP failures:

- Provider returns `429`.
- Tool request times out.

Example result:

```text
[FAIL] fallback_on_429: expected short_answer, got raw provider error
```

### 7.4 Verdict

Output:

- `SHIP`
- `SHIP WITH LIMITS`
- `HOLD`
- `BLOCK`

For the hackathon demo, focus on:

- `FinBank -> HOLD`
- `RetailCo -> SHIP`

This proves the core idea: customer-specific release gates.

## 8. Agent Harness Integration

Agent Sail should integrate with an agent harness through a simple target adapter.

MVP target types:

- `http`: call a running agent endpoint.
- `langgraph`: import a local LangGraph app/runnable and invoke it during criteria and load/SLO checks.
- `mock`: deterministic demo target.

Optional later:

- `command`: run a local command.
- `docker`: run an isolated local service.
- `a2a`: use remote specialist evaluators.

Minimal adapter contract:

```text
input: prompt, scenario metadata
output: answer, citations, actions, latency, error
```

For `langgraph`, the adapter should load a module path and app symbol, then normalize graph output into the same evidence contract.

Example target:

```bash
agentsail ci --customer finbank --target langgraph:./examples/support_graph.py:app
```

The harness should return structured evidence. Agent Sail should not rely on a model's subjective opinion for the final verdict.

## 9. CLI Commands

MVP:

```bash
agentsail ci --customer finbank --target http://localhost:8000/chat
agentsail ci --customer finbank --target langgraph:./examples/support_graph.py:app
agentsail ci --customer retailco --target http://localhost:8000/chat
agentsail report .agentsail/runs/finbank-run-001.json
```

Stretch:

```bash
agentsail learn context/acme/
agentsail tui
```

`learn` can turn customer notes into a draft criteria file, but it should not be required for the main demo. If used, it must show citations and allow human confirmation.

## 10. Demo Flow

Core demo in 3 minutes:

1. Show the same `Support Agent v12`.
2. Select `FinBank`.
3. Run `agentsail ci` against the LangGraph app target.
4. TUI shows live load/SLO probe.
5. Verdict: `HOLD`.
6. Failed criteria:
   - citations missing
   - only 31/50 users passed SLO
   - no fallback on 429
7. Select `RetailCo`.
8. Run the same agent against RetailCo criteria.
9. Verdict: `SHIP`.

One-line explanation:

> Same agent. Different customer. Different launch gate.

## 11. Demo Script

Opening:

> AI agents can build demos fast. The problem is that every customer defines success differently. Agent Sail turns customer-specific success criteria into release gates, then checks whether the agent is ready to ship.

During demo:

> FinBank needs citations, fallback behavior, and 50 concurrent internal users. The agent works in a demo, but Agent Sail blocks the launch because it fails FinBank's criteria.

Contrast:

> RetailCo has a different launch bar. Same agent, different customer, different verdict.

Closing:

> Agent Sail is customer-specific CI for AI agents.

## 12. Implementation Plan

12-hour scope:

| Time | Milestone |
|---|---|
| 0:00-1:00 | project setup, sample FastAPI target, sample LangGraph app, mock target |
| 1:00-2:00 | customer criteria YAML schema |
| 2:00-4:00 | criteria checks: citations, language, external action approval |
| 4:00-6:00 | load/SLO probe with `asyncio`, `httpx`, and LangGraph target adapter |
| 6:00-7:00 | chaos-lite 429/fallback check |
| 7:00-8:30 | verdict engine + run artifact JSON |
| 8:30-10:00 | TUI/live output with `rich.Live` |
| 10:00-11:00 | static HTML report |
| 11:00-12:00 | deterministic demo rehearsal |

Recommended libraries:

- `typer`
- `rich`
- `httpx`
- `pyyaml`
- `fastapi` for demo target
- `langgraph` for the demo app target adapter
- `jinja2` for static HTML report

Avoid for MVP:

- Full web app
- Real Slack/Notion/GitHub integrations
- Full A2A protocol
- LangGraph-only product positioning
- Production-grade load testing
- Real CI provider integration

## 13. Risks and Defenses

| Risk | Defense |
|---|---|
| "Isn't this just QA/CI?" | Traditional CI checks generic code behavior. Agent Sail checks customer-specific launch criteria for agent outputs. |
| "Can Codex just do this?" | Codex can fix code. Agent Sail defines repeatable release gates, runs probes, emits artifacts, and blocks release by policy. |
| "Is the load result production-accurate?" | It is a launch-readiness probe, not capacity certification. It catches SLO blockers before customer launch. |
| "Is the verdict subjective?" | Final verdict is rule-based from criteria and measured evidence. LLM extraction is optional and human-confirmed. |
| "Why not a web dashboard?" | The product is a CI gate. CLI/TUI is the source of truth; static HTML report is the shareable artifact. |

## 14. Roadmap

- Context extraction from meeting notes and customer docs.
- Human-confirmed customer contracts with citations.
- GitHub Actions integration.
- Historical verdict comparison.
- Replay failed customer scenarios.
- A2A specialist evaluators.
- Full web dashboard for account teams.
