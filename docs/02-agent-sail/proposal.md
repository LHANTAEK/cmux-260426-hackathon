# 02. Agent Sail — Customer-specific Agent Release Gate

## 1. One-liner

**Agent Sail turns customer-specific success criteria into executable, evidence-backed launch gates for AI agents.**

Short version:

> Before you show an agent to a customer, Agent Sail tells you: `SHIP`, `HOLD`, or `BLOCK` — with evidence.

## 2. Framing

Agent Sail is a **Developer Tooling** project, not a generic agent platform.

The product is not "one more multi-agent orchestrator." The product is the release gate around agent work:

- It automatically collects customer context from messenger threads, email, Notion, GitHub, issues, PRs, staging, and test results.
- It compiles that messy context into a customer-specific release contract.
- It runs deterministic checks, load/SLO probes, and chaos-lite failure probes.
- It emits a repeatable verdict and shareable evidence artifact.
- It blocks launch when the agent does not satisfy the customer contract.

Core message:

> AI agents can build demos fast. The problem is that every customer defines success differently. Agent Sail turns those customer-specific success criteria into executable release gates.

Do not pitch Agent Sail as a broad QA suite, SaaS dashboard, LLM gateway, or provider-specific assistant wrapper. Pitch it as a **customer-aware release verification harness**.

## 3. Problem

Small AI teams and agencies can build customer-facing agent demos quickly, but they often lack dedicated QA/QC capacity. The definition of "ready to launch" is fragmented across customer calls, PRDs, messenger threads, email, Slack messages, Notion pages, tickets, founder notes, GitHub issues, PRs, staging behavior, and test results.

That creates a release gap:

- The agent works in a demo.
- The customer has unstated or fragmented launch criteria.
- The team ships without checking those criteria as a repeatable gate.
- Bugs, missing requirements, tone drift, latency regressions, and unsafe fallbacks are discovered by the customer.

The same agent can be safe to ship for one customer and unsafe for another.

Example:

- `FinBank`: every answer needs citations, no PII exposure, 50 concurrent internal users, graceful fallback on provider `429`.
- `RetailCo`: Korean refund-policy answers, 20 concurrent users, fallback optional.

Agent Sail makes those customer-specific launch criteria executable.

## 4. Positioning

| Tool category | What it answers |
|---|---|
| Traditional CI | Does the code pass tests? |
| Observability | What happened after deployment? |
| LLM gateway | How should model requests be routed? |
| Load testing | How does the endpoint behave under traffic? |
| Generic QA tool | Does the product meet a broad checklist? |
| **Agent Sail** | **Does this agent satisfy this customer's launch criteria before launch?** |

Defense against "Codex/Claude can do this":

> Codex can fix code when asked. Agent Sail defines repeatable launch gates, runs the same probes every time, emits artifacts, and blocks release by policy.

## 5. Core Workflow

```text
Messenger / email / Notion / GitHub / PR / staging / tests
  -> context collectors
  -> customer contract
  -> target adapter
  -> criteria checks
  -> Playwright smoke checks
  -> load/SLO probe
  -> chaos-lite probe
  -> LLM omission/tone judge
  -> verdict engine
  -> exit code + rich TUI + cmux alert + report.json + report.html
  -> SHIP / HOLD / BLOCK
```

For the product, automatic context collection is the core. The source material is messenger, email, Notion, GitHub, PRs, staging, and tests. The compiled `customer_contract.json` is the normalized artifact that later checks consume.

For the hackathon, real SaaS integrations can be replaced with local fixtures that simulate messenger, email, Notion, and GitHub exports. Do not describe this as "manual criteria collection"; the MVP still demonstrates automatic ingest from scattered source material into a release contract.

The final verdict is rule-based from criteria and measurements, not subjective LLM opinion.

## 6. Product Surface

Agent Sail is CLI/TUI first, with a standalone HTML evidence report as the strongest shareable artifact.

Reasons:

- CMUX Hackathon is centered on terminal workflows and coding agents.
- Release gates need stdout, exit codes, and machine-readable artifacts.
- The load/SLO probe can be shown live in the terminal.
- CI integration is natural later.
- The HTML report gives customers, founders, and judges a concrete proof package.

Primary commands:

```bash
agentsail ci --customer finbank --target langgraph:./examples/support_graph.py:app
agentsail ci --customer retailco --target http://localhost:8000/chat
agentsail ci --customer finbank --target langgraph:./examples/support_graph.py:app --report --open
agentsail report .agentsail/runs/finbank-run-001.json
```

Generated artifacts:

```text
.agentsail/runs/finbank-run-001.json
.agentsail/contracts/finbank-contract.json
.agentsail/reports/finbank-run-001.html
```

The CLI is the source of truth. The TUI is the live release-board view. The HTML report is the evidence package. A cmux alert is a presentation amplifier, not the product body.

## 7. TUI Demo Surface

The TUI should be a release board, not a dashboard.

```text
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

## 8. HTML Evidence Report

The demo's strongest scene is an automatically opened HTML report with a clear top-line verdict.

Recommended structure:

1. `VERDICT`: `SHIP`, `HOLD`, or `BLOCK`.
2. `WHY`: the failed customer criteria and measured evidence.
3. `FIX NOW`: suggested patch, failing scenario, or next action.
4. `EVIDENCE`: criteria source, probe output, load metrics, chaos-lite result, screenshots/logs if available.
5. `PATCH`: generated patch sketch, failing Playwright test, or concrete next command when the issue is fixable.

Example top section:

```text
VERDICT: HOLD
Customer: FinBank
Target: langgraph:./examples/support_graph.py:app

Why held:
- Missing citations in 3/8 regulated-answer scenarios.
- Load probe reached only 31/50 users before p95 exceeded 5000ms.
- Provider 429 returned raw provider error instead of short fallback answer.
```

Important demo principle:

- The TUI shows the gate running.
- The HTML report proves why the gate blocked.
- `report.json` preserves the machine-readable verdict and evidence.
- The final story is not "we orchestrated agents." It is "we prevented a customer launch without evidence."

## 9. Automatic Context Collection

Agent Sail does not ask the team to manually type all launch criteria. It collects scattered customer context and compiles it into a release contract.

Default input sources:

- messenger threads: Slack, Discord, KakaoTalk, Teams, or exported chat logs
- email: customer asks, approvals, launch constraints, risk warnings
- Notion: PRD, customer notes, launch checklist, support policy
- GitHub: issues, PR descriptions, review comments, failing checks
- assistant logs: Claude/Codex/Gemini task logs and claimed completion summaries
- staging and tests: app URL, smoke-test results, Playwright traces

Hackathon implementation:

- Use local fixtures that look like messenger, email, Notion, GitHub, PR, and staging exports.
- Build `collectors/` that read those fixtures through the same interface real integrations would use.
- Compile them into `customer_contract.json`.
- Show the `context` workspace reading scattered inputs and producing 5-8 release contract items.

The important product claim is not "we store criteria YAML." The claim is:

> Agent Sail finds the launch criteria where the team already discussed them, turns them into a contract, and blocks release when the agent does not satisfy that contract.

## 10. Customer Contract Model

YAML can exist as an editable view or fixture format, but the product artifact is the compiled customer contract.

HTTP target contract:

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

LangGraph target contract:

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
    language: en
  reliability:
    expected_concurrency: 50
    max_p95_latency_ms: 5000
    max_error_rate: 0.01
    fallback_on_429: short_answer
  safety:
    pii_exposure: deny
    external_email_requires_approval: true
```

## 11. Target Adapters

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

The harness returns structured evidence. The verdict engine converts that evidence into `SHIP`, `HOLD`, or `BLOCK`.

## 12. MVP Checks

### Criteria Check

Deterministic response checks:

- citations exist when required
- response language matches customer contract
- external action includes approval metadata when required
- no obvious PII exposure when denied by customer policy
- required scenario outputs are present
- required customer deliverables from the compiled contract are present
- no drift from previously approved customer wording or launch constraints

### Playwright Smoke Check

Core user-flow checks:

- signup or auth path completes
- required post-signup email or notification fires
- mobile CTA and key customer-facing screens do not break
- required export or customer deliverable is visible

Example failure:

```text
[FAIL] user_flow: signup succeeds, but welcome email was not sent
```

### Load/SLO Probe

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

### Chaos-lite Probe

Inject predictable failures through the harness:

- provider `429`
- tool timeout
- empty retrieval result

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

When a failure is fixable, the report can include `PATCH` guidance: a patch diff, failing Playwright test, or concrete next command. `PATCH` is not a separate release approval state; it is an action lane inside `HOLD` or `BLOCK`.

For the demo:

- `FinBank -> HOLD`
- `RetailCo -> SHIP`

This proves the core product: same agent, different customer, different launch gate.

## 13. Hackathon Demo

3-minute flow:

1. Show `Support Agent v12`.
2. Show the `context` workspace reading messenger, email, Notion, GitHub, and PR fixtures into a compiled release contract.
3. Run `agentsail ci --customer finbank --target langgraph:./examples/support_graph.py:app --report --open`.
4. TUI shows criteria checks, Playwright smoke checks, and live load/SLO probe.
5. Verdict: `HOLD`.
6. Failed evidence: missing citations, only 31/50 users pass SLO, no `429` fallback.
7. Browser automatically opens `finbank-run-001.html`.
8. Report top line shows `VERDICT: HOLD` with failed criteria, measured evidence, and patch guidance.
9. GitHub/PR output shows merge blocked or check run failed.
10. Run the same agent for `RetailCo`.
11. Verdict: `SHIP`.

Talk track:

> Same agent. Different customer. Different launch gate.

Opening line:

> AI agents can build demos fast. The problem is that every customer defines success differently. Agent Sail turns those customer-specific success criteria into executable release gates.

Closing line:

> Code has CI before deployment. Customer-facing agents need Agent Sail before launch.

## 14. 12-hour Build Plan

| Time | Milestone |
|---|---|
| 0:00-1:00 | project setup, local messenger/email/Notion/GitHub fixtures, mock target, report template |
| 1:00-2:00 | `collectors/` fixture ingest and `customer_contract.json` compiler |
| 2:00-3:00 | HTTP and LangGraph target adapters |
| 3:00-4:30 | criteria checks: citations, language, external approval, PII deny, missing deliverables |
| 4:30-5:30 | Playwright smoke checks for core user flows |
| 5:30-7:00 | load/SLO probe with live metrics |
| 7:00-8:00 | chaos-lite `429`, timeout, empty retrieval checks |
| 8:00-9:00 | verdict engine, risk scoring, JSON artifact |
| 9:00-10:00 | TUI release board with `rich.Live` and cmux alert output |
| 10:00-11:00 | static HTML evidence report + `--open` + patch guidance |
| 11:00-12:00 | deterministic demo rehearsal and pitch cleanup |

Recommended libraries:

- `typer`
- `rich`
- `httpx`
- `pyyaml`
- `fastapi`
- `langgraph`
- `jinja2`
- `playwright`

Avoid for MVP:

- full web app
- real Slack/Notion/GitHub/email integrations
- full A2A protocol
- LangGraph-only positioning
- production-grade load testing
- real CI provider integration
- "supports every assistant" claims

## 15. Risks and Defenses

| Risk | Defense |
|---|---|
| "Isn't this just QA/CI?" | Traditional CI checks generic code behavior. Agent Sail checks customer-specific launch criteria for agent outputs. |
| "Can Codex just do this?" | Codex can fix code. Agent Sail makes the criteria repeatable, measurable, enforceable, and artifact-backed. |
| "Is the load result production-accurate?" | It is a launch-readiness probe, not a capacity certification. It catches obvious SLO blockers before customer launch. |
| "Is the verdict subjective?" | Final verdict is rule-based from customer criteria and measured evidence. |
| "Why not a web dashboard?" | The product is a release gate. CLI/TUI is the source of truth; HTML is the shareable evidence artifact. cmux is a demo amplifier. |
| "Is this a generic agent platform?" | No. It is a release verification harness focused on customer launch readiness. |
| "Are criteria manually entered?" | No. Agent Sail automatically collects scattered context from messenger, email, Notion, GitHub, PRs, staging, and tests, then compiles a release contract. |

## 16. Roadmap

- Production connectors for Slack, Teams, Gmail, Notion, Linear, GitHub, and PR comments.
- Human-confirmed customer criteria with citations.
- GitHub Actions integration and blocking check runs.
- Historical verdict comparison.
- Failed scenario replay.
- Account-team web dashboard.
- A2A specialist evaluators for narrow criteria checks.

## 17. Session Memory

Keep these constraints intact in future edits:

- Agent Sail is a **customer-aware release gate**, not a generic orchestration platform.
- The source material is scattered customer context: messenger, email, Notion, GitHub, PRs, staging, tests, and assistant logs.
- The compiled artifact is `customer_contract.json`; the flow is `collected context -> release contract -> measured evidence -> verdict`.
- The CLI owns exit codes, `report.json`, and machine-readable artifacts.
- The TUI is the live release-board view.
- cmux is a visual amplifier, not the product body.
- The HTML report is the shareable evidence package and the strongest demo artifact.
- The MVP should use local fixtures for messenger/email/Notion/GitHub ingest, not manual criteria entry.
- The key demo contrast is `FinBank -> HOLD` and `RetailCo -> SHIP` for the same agent.
- Alternate fixed demo values from the source framing: `acme-bank`, failures `missing CSV export`, `beta badge exposed`, `tone drift`, then cmux red alert and `report.html` auto-open.
- The strongest line is: **Same agent. Different customer. Different launch gate.**

## 18. README3 Merge Check

This section records exactly how the important `README3.md` ideas were merged into Agent Sail, and what was intentionally renamed or narrowed.

| README3 point | Agent Sail treatment | Where it appears |
|---|---|---|
| Developer Tooling, not AI Safety | Preserved. Agent Sail is framed as developer tooling for customer-facing agent release verification. | Sections 2, 4, 17 |
| Not a generic multi-agent platform | Preserved. The product is the release gate around agent work, not orchestration itself. | Sections 2, 15, 17 |
| Assistant-agnostic wrapper over Claude/Codex/Gemini/custom pipelines | Preserved as a gate that can wrap assistant logs and target adapters without claiming every assistant is fully implemented in the MVP. | Sections 9, 11, 17 |
| Scattered source material: Slack, Notion, GitHub, agent conversation | Expanded to messenger, email, Notion, GitHub, PRs, staging, tests, and assistant logs. | Sections 3, 5, 9, 17 |
| Automatic context collection, not manual criteria typing | Corrected. YAML is only an editable view or fixture format; the core flow is automatic ingest into `customer_contract.json`. | Sections 5, 9, 10, 14, 17 |
| Release contract auto-generation | Preserved as `customer_contract.json` compilation before checks run. | Sections 5, 9, 10, 14, 17 |
| Verify requirements omission, user-flow bugs, and drift | Preserved. Checks include missing deliverables, Playwright smoke, prior-agreement drift, and LLM omission/tone judge. | Sections 5, 12, 18 |
| Project context, PR, staging, and test results are all inputs | Preserved and made explicit in the workflow and collectors. | Sections 5, 9 |
| Standalone HTML evidence report plus auto-open is the demo star | Preserved. HTML report remains the strongest shareable artifact. | Sections 6, 8, 13, 17 |
| TUI/GUI are optional channels | Preserved. TUI is the live board; cmux is a visual amplifier; HTML is the proof package. | Sections 6, 7, 8, 17 |
| Output includes exit code, rich TUI, cmux alert, `report.json`, `report.html` | Preserved. The generated artifact list now includes `customer_contract.json` too. | Sections 5, 6, 14, 17 |
| Final action is allow/block | Mapped to Agent Sail language as `SHIP`, `HOLD`, `BLOCK`; `PATCH` is represented as a guidance lane inside `HOLD` or `BLOCK`. | Sections 8, 12 |
| PR/merge blocking matters | Preserved as demo output and future GitHub Actions/check-run integration. | Sections 13, 16 |
| Product body is verdict engine; cmux is not the body | Preserved explicitly. | Sections 6, 15, 17 |
| Do not overbuild a full SaaS dashboard | Preserved in MVP exclusions and roadmap boundaries. | Sections 14, 15, 17 |
| Demo fixed values: `acme-bank`, `missing CSV export`, `beta badge exposed`, `tone drift` | Preserved as alternate fixed demo values while the primary Agent Sail example keeps `FinBank -> HOLD` and `RetailCo -> SHIP`. | Section 17 |

No important README3 product constraint is intentionally dropped. The only meaningful rename is from `shiplock-harness` language to Agent Sail language:

- `release contract` -> `customer_contract.json`
- `PASS / BLOCK / PATCH` -> `SHIP / HOLD / BLOCK` plus `PATCH` guidance
- `Slack + Notion + GitHub` -> messenger/email/Notion/GitHub/PR/staging/tests/assistant logs
- `cmux warning` -> cmux alert as presentation amplifier
