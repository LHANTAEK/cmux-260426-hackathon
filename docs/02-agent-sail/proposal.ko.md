# 02. Agent Sail — 고객별 에이전트 출시 게이트

## 1. 한 줄 요약

**Agent Sail은 고객별 성공 기준을 실행 가능한, 증거 기반의 AI 에이전트 출시 게이트로 바꾼다.**

짧은 버전:

> 고객에게 에이전트를 보여주기 전에, Agent Sail이 `SHIP`, `HOLD`, `BLOCK`을 증거와 함께 판정한다.

## 2. 프레이밍

Agent Sail은 **Developer Tooling** 프로젝트다. 범용 에이전트 플랫폼이 아니다.

제품의 본체는 "또 하나의 멀티에이전트 오케스트레이터"가 아니라, 에이전트 작업물 바깥을 감싸는 출시 게이트다.

- 메신저 thread, 메일, Notion, GitHub, issue, PR, staging, test result에서 고객 맥락을 자동 수집한다.
- 흩어진 맥락을 고객별 release contract로 컴파일한다.
- 결정론적 기준 검사, load/SLO probe, chaos-lite 실패 probe를 실행한다.
- 반복 가능한 verdict와 공유 가능한 evidence artifact를 남긴다.
- 에이전트가 고객 계약을 만족하지 못하면 출시를 막는다.

핵심 메시지:

> AI 에이전트는 데모를 빠르게 만들 수 있다. 문제는 고객마다 성공 기준이 다르다는 것이다. Agent Sail은 그 고객별 성공 기준을 실행 가능한 출시 게이트로 바꾼다.

Agent Sail을 광범위한 QA suite, SaaS dashboard, LLM gateway, 특정 provider용 assistant wrapper로 설명하지 않는다. **고객 맥락을 아는 release verification harness**로 설명한다.

## 3. 문제

작은 AI 팀과 에이전시는 고객-facing 에이전트 데모를 빠르게 만들 수 있지만, 전담 QA/QC 인력을 두기 어렵다. "출시해도 되는 상태"의 정의는 고객 미팅, PRD, 메신저 thread, 메일, Slack 메시지, Notion 페이지, 티켓, founder note, GitHub issue, PR, staging 동작, test result에 흩어져 있다.

그래서 출시 공백이 생긴다.

- 에이전트는 데모에서는 동작한다.
- 고객의 출시 기준은 명시되지 않았거나 여러 곳에 쪼개져 있다.
- 팀은 그 기준을 반복 가능한 게이트로 검증하지 않고 출고한다.
- 버그, 누락 요구사항, tone drift, latency regression, 위험한 fallback을 고객이 먼저 발견한다.

같은 에이전트도 어떤 고객에게는 출시 가능하고, 다른 고객에게는 출시하면 안 될 수 있다.

예시:

- `FinBank`: 모든 답변에 citation 필요, PII 노출 금지, 내부 사용자 50명 동시 사용, provider `429` 발생 시 graceful fallback 필요.
- `RetailCo`: 한국어 환불 정책 답변, 동시 사용자 20명, fallback은 선택 사항.

Agent Sail은 이런 고객별 출시 기준을 실행 가능하게 만든다.

## 4. 포지셔닝

| 도구 분류 | 답하는 질문 |
|---|---|
| 전통적 CI | 코드 테스트가 통과하는가? |
| Observability | 배포 후 무슨 일이 일어났는가? |
| LLM Gateway | 모델 요청을 어떻게 라우팅할 것인가? |
| Load Testing | 트래픽 아래 endpoint가 어떻게 동작하는가? |
| Generic QA Tool | 제품이 일반 체크리스트를 만족하는가? |
| **Agent Sail** | **이 에이전트가 이 고객의 출시 기준을 출시 전에 만족하는가?** |

"Codex/Claude가 하면 되는 것 아닌가?"에 대한 방어:

> Codex는 요청받으면 코드를 고칠 수 있다. Agent Sail은 출시 기준을 반복 가능하고, 측정 가능하고, 강제 가능하고, artifact로 남는 게이트로 만든다.

## 5. 핵심 워크플로우

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

제품의 핵심은 자동 context collection이다. Source material은 메신저, 메일, Notion, GitHub, PR, staging, test다. 컴파일된 `customer_contract.json`이 이후 check들이 소비하는 normalized artifact다.

해커톤에서는 실제 SaaS integration 대신 메신저, 메일, Notion, GitHub export처럼 보이는 local fixture를 쓴다. 이것을 "수동 criteria 입력"이라고 설명하면 안 된다. MVP도 흩어진 source material을 자동 ingest해서 release contract로 만드는 장면을 보여줘야 한다.

최종 verdict는 주관적인 LLM 판단이 아니라, criteria와 measurement에서 나온 rule-based 판정이다.

## 6. 제품 표면

Agent Sail은 CLI/TUI first 제품이고, standalone HTML evidence report가 가장 강한 공유 artifact다.

이유:

- CMUX Hackathon은 terminal workflow와 coding agent 중심이다.
- 출시 게이트에는 stdout, exit code, machine-readable artifact가 필요하다.
- load/SLO probe를 terminal에서 live로 보여줄 수 있다.
- 나중에 CI integration으로 자연스럽게 이어진다.
- HTML report는 고객, founder, 심사위원에게 보여줄 수 있는 구체적인 증거 패키지다.

주요 명령:

```bash
agentsail ci --customer finbank --target langgraph:./examples/support_graph.py:app
agentsail ci --customer retailco --target http://localhost:8000/chat
agentsail ci --customer finbank --target langgraph:./examples/support_graph.py:app --report --open
agentsail report .agentsail/runs/finbank-run-001.json
```

생성 artifact:

```text
.agentsail/runs/finbank-run-001.json
.agentsail/contracts/finbank-contract.json
.agentsail/reports/finbank-run-001.html
```

CLI가 source of truth다. TUI는 live release-board view다. HTML report는 evidence package다. cmux alert는 발표 증폭 장치이지 제품 본체가 아니다.

## 7. TUI 데모 화면

TUI는 dashboard가 아니라 release board여야 한다.

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

안정성을 위해 `rich.Live`로 구현한다. Textual은 시간이 남을 때만 선택한다.

## 8. HTML Evidence Report

데모에서 가장 강한 장면은 자동으로 열린 HTML report에 top-line verdict가 명확하게 보이는 것이다.

권장 구조:

1. `VERDICT`: `SHIP`, `HOLD`, `BLOCK`
2. `WHY`: 실패한 고객 기준과 측정 증거
3. `FIX NOW`: 제안 patch, 실패 scenario, 다음 action
4. `EVIDENCE`: criteria source, probe output, load metrics, chaos-lite result, 필요하면 screenshot/log
5. `PATCH`: 고칠 수 있는 문제일 때 generated patch sketch, failing Playwright test, concrete next command

상단 예시:

```text
VERDICT: HOLD
Customer: FinBank
Target: langgraph:./examples/support_graph.py:app

Why held:
- 규제 답변 scenario 8개 중 3개에서 citation 누락.
- load probe가 50명 목표 전에 p95 5000ms를 초과했고 31명까지만 통과.
- provider 429에서 short fallback answer 대신 raw provider error 반환.
```

중요한 데모 원칙:

- TUI는 게이트가 실행되는 모습을 보여준다.
- HTML report는 왜 막혔는지를 증명한다.
- `report.json`은 machine-readable verdict와 evidence를 보존한다.
- 최종 스토리는 "에이전트를 orchestrate했다"가 아니라 "증거 없는 고객 출시를 막았다"다.

## 9. 자동 Context Collection

Agent Sail은 팀에게 출시 기준을 전부 손으로 입력하라고 하지 않는다. 흩어진 고객 맥락을 수집하고 release contract로 컴파일한다.

기본 입력 source:

- messenger thread: Slack, Discord, KakaoTalk, Teams, exported chat log
- email: 고객 요청, 승인, 출시 제약, risk warning
- Notion: PRD, customer note, launch checklist, support policy
- GitHub: issue, PR description, review comment, failing check
- assistant log: Claude/Codex/Gemini 작업 로그와 완료 주장
- staging and tests: app URL, smoke-test result, Playwright trace

해커톤 구현:

- 메신저, 메일, Notion, GitHub, PR, staging export처럼 보이는 local fixture를 사용한다.
- 실제 integration과 같은 interface를 쓰는 `collectors/`를 만든다.
- 수집 결과를 `customer_contract.json`으로 컴파일한다.
- `context` workspace에서 흩어진 입력을 읽고 release contract item 5-8개를 생성하는 장면을 보여준다.

중요한 제품 주장은 "criteria YAML을 저장한다"가 아니다. 핵심은 이것이다.

> Agent Sail은 팀이 이미 논의한 곳에서 출시 기준을 찾아 contract로 바꾸고, 에이전트가 그 contract를 만족하지 못하면 release를 막는다.

## 10. Customer Contract Model

YAML은 editable view나 fixture format으로 존재할 수 있다. 하지만 제품 artifact는 compiled customer contract다.

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

Agent Sail은 target adapter를 통해 에이전트를 검사한다.

MVP adapter:

- `langgraph`: local LangGraph app/runnable을 import해서 criteria check와 load/SLO check 중 invoke한다.
- `http`: 실행 중인 agent endpoint를 `httpx`로 호출한다.
- `mock`: 안정적인 발표를 위한 deterministic demo target.

Adapter contract:

```text
input: prompt, scenario metadata
output: answer, citations, actions, latency, error
```

harness는 structured evidence를 반환한다. verdict engine은 그 evidence를 `SHIP`, `HOLD`, `BLOCK`으로 변환한다.

## 12. MVP Checks

### Criteria Check

결정론적 response check:

- required일 때 citation이 존재하는가
- 응답 언어가 customer contract와 맞는가
- 외부 action에 approval metadata가 포함되는가
- 고객 policy가 금지한 PII 노출이 없는가
- required scenario output이 존재하는가
- compiled contract에 포함된 필수 customer deliverable이 존재하는가
- 기존 고객 합의 문구나 출시 제약에서 drift가 없는가

### Playwright Smoke Check

핵심 사용자 플로우 검사:

- signup 또는 auth path가 완료되는가
- signup 이후 필요한 email 또는 notification이 발송되는가
- mobile CTA와 주요 customer-facing 화면이 깨지지 않는가
- required export 또는 customer deliverable이 보이는가

실패 예시:

```text
[FAIL] user_flow: signup succeeds, but welcome email was not sent
```

### Load/SLO Probe

production capacity certification이 아니라 짧은 launch-readiness probe다.

MVP:

- `asyncio` load driver
- HTTP target에는 `httpx`
- LangGraph target에는 direct invocation
- expected concurrency까지 짧은 ramp
- p50, p95, error rate, completed users 수집

실패 예시:

```text
[FAIL] load: reached only 31/50 users before p95 exceeded 5000ms
```

### Chaos-lite Probe

harness를 통해 예측 가능한 실패를 주입한다.

- provider `429`
- tool timeout
- empty retrieval result

실패 예시:

```text
[FAIL] fallback_on_429: expected short_answer, got raw provider error
```

### Verdict

가능한 출력:

- `SHIP`
- `SHIP WITH LIMITS`
- `HOLD`
- `BLOCK`

고칠 수 있는 실패라면 report가 `PATCH` guidance를 포함할 수 있다. 예: patch diff, failing Playwright test, concrete next command. `PATCH`는 별도 release approval state가 아니라 `HOLD` 또는 `BLOCK` 안의 action lane이다.

데모 고정값:

- `FinBank -> HOLD`
- `RetailCo -> SHIP`

이 대비가 제품의 핵심을 증명한다. 같은 에이전트, 다른 고객, 다른 출시 게이트.

## 13. 해커톤 데모

3분 플로우:

1. `Support Agent v12`를 보여준다.
2. `context` workspace가 메신저, 메일, Notion, GitHub, PR fixture를 읽어 compiled release contract를 만드는 장면을 보여준다.
3. `agentsail ci --customer finbank --target langgraph:./examples/support_graph.py:app --report --open` 실행.
4. TUI가 criteria check, Playwright smoke check, live load/SLO probe를 보여준다.
5. Verdict: `HOLD`.
6. 실패 evidence: citation 누락, 50명 중 31명까지만 SLO 통과, `429` fallback 없음.
7. 브라우저가 `finbank-run-001.html`을 자동으로 연다.
8. Report 상단에 `VERDICT: HOLD`, 실패 criteria, 측정 evidence, patch guidance가 보인다.
9. GitHub/PR output에 merge blocked 또는 check run failed가 보인다.
10. 같은 에이전트를 `RetailCo`로 실행한다.
11. Verdict: `SHIP`.

Talk track:

> Same agent. Different customer. Different launch gate.

Opening line:

> AI agents can build demos fast. The problem is that every customer defines success differently. Agent Sail turns those customer-specific success criteria into executable release gates.

Closing line:

> Code has CI before deployment. Customer-facing agents need Agent Sail before launch.

## 14. 12시간 빌드 플랜

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

권장 라이브러리:

- `typer`
- `rich`
- `httpx`
- `pyyaml`
- `fastapi`
- `langgraph`
- `jinja2`
- `playwright`

MVP에서 피할 것:

- full web app
- 실제 Slack/Notion/GitHub/email integration
- full A2A protocol
- LangGraph-only positioning
- production-grade load testing
- 실제 CI provider integration
- "모든 assistant 지원" 같은 과장

## 15. 리스크와 방어

| Risk | Defense |
|---|---|
| "그냥 QA/CI 아닌가?" | 전통적 CI는 generic code behavior를 본다. Agent Sail은 customer-specific launch criteria를 agent output 기준으로 검사한다. |
| "Codex가 하면 되는 것 아닌가?" | Codex는 코드를 고칠 수 있다. Agent Sail은 기준을 반복 가능하고, 측정 가능하고, 강제 가능하고, artifact-backed한 게이트로 만든다. |
| "Load 결과가 production-accurate한가?" | capacity certification이 아니라 launch-readiness probe다. 고객 출시 전 명백한 SLO blocker를 잡는 것이 목적이다. |
| "Verdict가 주관적인가?" | 최종 verdict는 customer criteria와 measured evidence에서 나온 rule-based 판정이다. |
| "왜 web dashboard가 아닌가?" | 제품은 release gate다. CLI/TUI가 source of truth이고 HTML은 shareable evidence artifact다. cmux는 demo amplifier다. |
| "범용 agent platform인가?" | 아니다. customer launch readiness에 집중한 release verification harness다. |
| "Criteria를 손으로 입력하는가?" | 아니다. Agent Sail은 메신저, 메일, Notion, GitHub, PR, staging, test에서 흩어진 맥락을 자동 수집한 뒤 release contract로 컴파일한다. |

## 16. Roadmap

- Slack, Teams, Gmail, Notion, Linear, GitHub, PR comment용 production connector.
- citation이 붙은 human-confirmed customer criteria.
- GitHub Actions integration과 blocking check run.
- 과거 verdict 비교.
- 실패 scenario replay.
- Account-team web dashboard.
- 좁은 criteria check를 위한 A2A specialist evaluator.

## 17. 세션 메모리

이후 편집에서도 반드시 유지할 제약:

- Agent Sail은 **고객 맥락을 아는 release gate**다. 범용 orchestration platform이 아니다.
- Source material은 메신저, 메일, Notion, GitHub, PR, staging, test, assistant log에 흩어진 고객 맥락이다.
- Compiled artifact는 `customer_contract.json`이고, 흐름은 `collected context -> release contract -> measured evidence -> verdict`다.
- CLI가 exit code, `report.json`, machine-readable artifact를 책임진다.
- TUI는 live release-board view다.
- cmux는 visual amplifier이지 제품 본체가 아니다.
- HTML report는 shareable evidence package이며 가장 강한 데모 artifact다.
- MVP는 manual criteria entry가 아니라 메신저/메일/Notion/GitHub ingest용 local fixture를 사용한다.
- 핵심 데모 대비는 같은 에이전트에 대해 `FinBank -> HOLD`, `RetailCo -> SHIP`이다.
- 원본 프레이밍의 alternate demo fixed values: `acme-bank`, 실패 `missing CSV export`, `beta badge exposed`, `tone drift`, 이후 cmux red alert와 `report.html` auto-open.
- 가장 강한 문장은 **Same agent. Different customer. Different launch gate.**다.

## 18. README3 병합 체크

이 섹션은 `README3.md`의 중요사항을 Agent Sail 문서에 어떻게 병합했는지, 무엇을 이름만 바꿨는지, 무엇을 의도적으로 좁혔는지 기록한다.

| README3 중요사항 | Agent Sail 반영 방식 | 반영 위치 |
|---|---|---|
| AI Safety가 아니라 Developer Tooling | 유지. 고객-facing agent의 release verification을 위한 developer tool로 포지셔닝했다. | 2, 4, 17 |
| 범용 multi-agent platform이 아님 | 유지. 제품 본체는 orchestration이 아니라 agent 작업물 바깥의 release gate다. | 2, 15, 17 |
| Claude/Codex/Gemini/custom pipeline을 감싸는 assistant-agnostic wrapper | 유지. MVP에서 모든 assistant를 구현한다고 과장하지 않고, assistant log와 target adapter를 감싸는 gate로 정리했다. | 9, 11, 17 |
| Slack, Notion, GitHub, agent conversation에 흩어진 원자료 | 확장. messenger, email, Notion, GitHub, PR, staging, test, assistant log를 source material로 명시했다. | 3, 5, 9, 17 |
| 수동 criteria 입력이 아니라 자동 context collection | 수정 완료. YAML은 editable view나 fixture format일 뿐이고, 핵심은 자동 ingest 후 `customer_contract.json` 생성이다. | 5, 9, 10, 14, 17 |
| release contract 자동 생성 | 유지. Check 전에 `customer_contract.json`으로 컴파일하는 단계로 반영했다. | 5, 9, 10, 14, 17 |
| 요구사항 누락, user-flow bug, 기존 합의와의 drift 검증 | 유지. Missing deliverable, Playwright smoke, prior-agreement drift, LLM omission/tone judge를 check 범위에 넣었다. | 5, 12, 18 |
| project context, PR, staging, test result를 한 번에 입력으로 사용 | 유지. Workflow와 collectors 입력 source에 명시했다. | 5, 9 |
| standalone HTML evidence report + auto-open이 데모 주인공 | 유지. HTML report를 가장 강한 shareable artifact로 유지했다. | 6, 8, 13, 17 |
| TUI/GUI는 부가 채널 | 유지. TUI는 live board, cmux는 visual amplifier, HTML은 proof package로 구분했다. | 6, 7, 8, 17 |
| 출력은 exit code + rich TUI + cmux alert + report.json + report.html | 유지. 여기에 `customer_contract.json` artifact도 추가했다. | 5, 6, 14, 17 |
| 최종 action은 allow/block | Agent Sail 언어로 `SHIP`, `HOLD`, `BLOCK`에 매핑했다. `PATCH`는 `HOLD`/`BLOCK` 내부의 guidance lane으로 보존했다. | 8, 12 |
| PR/merge blocking이 중요 | 유지. 데모 output과 roadmap의 GitHub Actions/check-run integration에 반영했다. | 13, 16 |
| 제품 본체는 verdict engine이고 cmux는 본체가 아님 | 명시적으로 유지했다. | 6, 15, 17 |
| full SaaS dashboard로 확장하지 말 것 | MVP exclusion과 roadmap boundary에 반영했다. | 14, 15, 17 |
| demo fixed values: `acme-bank`, `missing CSV export`, `beta badge exposed`, `tone drift` | primary 예시는 `FinBank -> HOLD`, `RetailCo -> SHIP`로 유지하되, 원본 프레이밍의 alternate fixed demo values로 보존했다. | 17 |

의도적으로 버린 핵심 제약은 없다. 이름만 Agent Sail 문맥에 맞게 바꾼 부분은 아래다.

- `release contract` -> `customer_contract.json`
- `PASS / BLOCK / PATCH` -> `SHIP / HOLD / BLOCK` + `PATCH` guidance
- `Slack + Notion + GitHub` -> messenger/email/Notion/GitHub/PR/staging/tests/assistant logs
- `cmux warning` -> presentation amplifier로서의 cmux alert
