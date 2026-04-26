# 02. Agent Sail — Customer-specific Agent CI

## 1. One-liner

**Agent Sail turns customer-specific success criteria into executable launch gates for AI agents.**

Short version:

> Before you show an agent to a customer, Agent Sail tells you: `SHIP`, `HOLD`, or `BLOCK`.

## 2. Problem

AI teams can now build agent demos quickly with coding agents, but the definition of "success" is usually scattered across customer calls, PRDs, Slack messages, tickets, and founder notes.

That creates a release gap:

- The agent works in a demo.
- The customer has unstated or fragmented launch criteria.
- The team ships without checking those criteria as a repeatable gate.

The same agent can be safe to ship for one customer and unsafe for another.

Example:

- `FinBank`: every answer needs citations, no PII exposure, 50 concurrent internal users, graceful fallback on provider `429`.
- `RetailCo`: Korean refund-policy answers, 20 concurrent users, fallback optional.

Agent Sail makes those customer-specific launch criteria executable.

## 3. Positioning

| Tool category | What it answers |
|---|---|
| Traditional CI | Does the code pass tests? |
| Observability | What happened after deployment? |
| LLM gateway | How should model requests be routed? |
| Load testing | How does the endpoint behave under traffic? |
| **Agent Sail** | **Does this agent satisfy this customer's launch criteria?** |

Agent Sail is not a generic QA tool, observability dashboard, or gateway. It is a pre-launch release gate for customer-facing agents.

Defense against "Codex/Claude can do this":

> Codex can fix code when asked. Agent Sail defines repeatable launch gates, runs the same probes every time, emits artifacts, and blocks release by policy.

## 4. Core Workflow

```
Customer context
  -> customer criteria YAML
  -> agent harness checks
  -> load/SLO probe
  -> chaos-lite probe
  -> SHIP / HOLD / BLOCK artifact
```

For the hackathon, explicit YAML criteria are the source of truth. Context extraction from messy notes is a stretch goal, not the MVP.

## 5. Primary Interface

Agent Sail should be CLI/TUI first.

Reasons:

- CMUX Hackathon is centered on terminal workflows and coding agents.
- Release gates need stdout, exit codes, and artifacts.
- The load/SLO probe can be shown live in the terminal.
- CI integration is natural later.

Primary commands:

```bash
agentsail ci --customer finbank --target langgraph:./examples/support_graph.py:app
agentsail ci --customer retailco --target http://localhost:8000/chat
agentsail report .agentsail/runs/finbank-run-001.json
```

Static HTML is secondary:

```bash
agentsail ci --customer finbank --target langgraph:./examples/support_graph.py:app --report
```

Generated artifact:

```text
.agentsail/reports/finbank-run-001.html
```

## 6. TUI Demo Surface

The TUI should be a release board, not a dashboard.

```
┌ Customers ───────┐ ┌ Criteria / Evidence ─────────────────────┐ ┌ Verdict ───────┐
│ FinBank          │ │ citations_required                        │ │ HOLD            │
│ RetailCo         │ │ expected_concurrency: 50                  │ │                 │
│                  │ │ max_p95_latency_ms: 5000                  │ │ Failed:         │
│ Target           │ │ fallback_on_429: short_answer             │ │ 1. citations    │
│ langgraph:app    │ │ source: finbank.yaml                      │ │ 2. 31/50 users  │
└──────────────────┘ └───────────────────────────────────────────┘ └─────────────────┘

┌ Live Load Probe ───────────────────────────────────────────────────────────────┐
│ Phase: RAMPING     Users: 37 / 50     RPS: 12.4                               │
│ p50: 1.2s          p95: 5.8s          error: 0.4%                              │
│ Load      ▁▂▃▄▅▆▇                                                              │
│ p95       ▁▁▂▂▄▆█                                                              │
└────────────────────────────────────────────────────────────────────────────────┘
```

Implement with `rich.Live` for reliability. Textual is optional only if time remains.

## 7. Customer Criteria Model

HTTP target:

```yaml
customer: retailco
target:
  type: http
  endpoint: http://localhost:8000/chat
  method: POST
  prompt_field: message
  answer_field: answer

criteria:
  quality:
    language: ko
  reliability:
    expected_concurrency: 20
    max_p95_latency_ms: 5000
    max_error_rate: 0.01
```

LangGraph target:

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

## 8. Agent Harness

Agent Sail checks an agent through target adapters.

MVP adapters:

- `langgraph`: import a local LangGraph app/runnable and invoke it during criteria and load/SLO checks.
- `http`: call a running agent endpoint with `httpx`.
- `mock`: deterministic demo target for reliable presentation.

Adapter contract:

```text
input: prompt, scenario metadata
output: answer, citations, actions, latency, error
```

The harness returns structured evidence. Final verdicts are rule-based from criteria and measurements, not subjective LLM opinions.

## 9. MVP Checks

### Criteria check

Deterministic response checks:

- citations exist when required
- response language matches customer contract
- external action includes approval metadata when required

### Load/SLO probe

Short launch-readiness probe, not production capacity certification.

MVP:

- `asyncio` load driver
- `httpx` for HTTP targets
- direct invocation for LangGraph targets
- short ramp to expected concurrency
- collect p50, p95, error rate, completed users

Example failure:

```text
[FAIL] load: reached only 31/50 users before p95 exceeded 5000ms
```

### Chaos-lite probe

Inject predictable failures through the harness:

- provider `429`
- tool timeout

Example failure:

```text
[FAIL] fallback_on_429: expected short_answer, got raw provider error
```

### Verdict

Possible outputs:

- `SHIP`
- `SHIP WITH LIMITS`
- `HOLD`
- `BLOCK`

For the demo:

- `FinBank -> HOLD`
- `RetailCo -> SHIP`

This proves the core product: same agent, different customer, different launch gate.

## 10. Hackathon Demo

3-minute flow:

1. Show `Support Agent v12`.
2. Run `agentsail ci --customer finbank --target langgraph:./examples/support_graph.py:app`.
3. TUI shows criteria checks and live load/SLO probe.
4. Verdict: `HOLD`.
5. Failed evidence: missing citations, only 31/50 users pass SLO, no `429` fallback.
6. Run the same agent for `RetailCo`.
7. Verdict: `SHIP`.

Talk track:

> Same agent. Different customer. Different launch gate.

Opening line:

> AI agents can build demos fast. The problem is that every customer defines success differently. Agent Sail turns those customer-specific success criteria into executable release gates.

Closing line:

> Code has CI before deployment. Customer-facing agents need Agent Sail before launch.

## 11. 12-hour Build Plan

| Time | Milestone |
|---|---|
| 0:00-1:00 | project setup, demo criteria files, mock target |
| 1:00-2:00 | HTTP and LangGraph target adapters |
| 2:00-3:30 | criteria checks: citations, language, external approval |
| 3:30-5:30 | load/SLO probe with live metrics |
| 5:30-6:30 | chaos-lite `429` and timeout checks |
| 6:30-7:30 | verdict engine and JSON artifact |
| 7:30-9:00 | TUI with `rich.Live` |
| 9:00-10:00 | static HTML report |
| 10:00-12:00 | deterministic demo rehearsal and pitch cleanup |

Recommended libraries:

- `typer`
- `rich`
- `httpx`
- `pyyaml`
- `fastapi`
- `langgraph`
- `jinja2`

Avoid for MVP:

- full web app
- Slack/Notion/GitHub integrations
- full A2A protocol
- LangGraph-only positioning
- production-grade load testing
- real CI provider integration

## 12. Risks and Defenses

| Risk | Defense |
|---|---|
| "Isn't this just QA/CI?" | Traditional CI checks generic code behavior. Agent Sail checks customer-specific launch criteria for agent outputs. |
| "Can Codex just do this?" | Codex can fix code. Agent Sail makes the criteria repeatable, measurable, and enforceable. |
| "Is the load result production-accurate?" | It is a launch-readiness probe, not a capacity certification. It catches obvious SLO blockers before customer launch. |
| "Is the verdict subjective?" | Final verdict is rule-based from customer criteria and measured evidence. |
| "Why not a web dashboard?" | The product is a release gate. CLI/TUI is the source of truth; HTML is only a shareable artifact. |

## 13. Roadmap

- Context extraction from meeting notes and customer docs.
- Human-confirmed customer criteria with citations.
- GitHub Actions integration.
- Historical verdict comparison.
- Failed scenario replay.
- A2A specialist evaluators.
- Account-team web dashboard.
