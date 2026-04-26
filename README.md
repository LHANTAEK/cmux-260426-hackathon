# Agent Sail

> **Same agent. Different customer. Different launch gate.**

Agent Sail은 고객별 성공 기준을 실행 가능한, 증거 기반의 AI 에이전트 출시 게이트로 바꾼다.
고객에게 에이전트를 보여주기 전에, Agent Sail이 `SHIP`, `HOLD`, `BLOCK`을 evidence와 함께 판정한다.

설계 원본은 [`docs/02-agent-sail/proposal.ko.md`](docs/02-agent-sail/proposal.ko.md).

---

## 무엇을 푸는가

작은 AI 팀과 에이전시는 데모를 빠르게 만들지만, "출시해도 되는 상태"의 정의가
고객 미팅 / Slack / 메일 / Notion / PRD / GitHub PR / staging 동작에 흩어져 있다.
같은 에이전트도 어떤 고객에게는 출시 가능하고, 다른 고객에게는 출시하면 안 된다.

| 고객 | 기준 | 데모 verdict |
|---|---|---|
| `finbank` | 모든 답변에 citation, PII 노출 금지, 50명 동시, `429` graceful fallback | **HOLD** |
| `retailco` | 한국어 환불 정책 답변, 20명 동시, fallback 선택 | **SHIP** |
| `acme-bank` | enterprise tone, white-label only, CSV export 필수, beta 배지 금지 | **BLOCK** |

`acme-bank`의 BLOCK 사유는 `missing CSV export`, `beta badge exposed`, `tone drift` 세 줄이다.
이 대비가 제품의 본체다.

---

## 핵심 워크플로우

```text
Messenger / email / Notion / GitHub / PR / staging / tests
  -> context collectors          (.agentsail/cache/<customer>/*.json)
  -> customer contract           (.agentsail/contracts/<customer>-contract.json)
  -> target adapter              (langgraph: / http: / mock:)
  -> criteria checks             (citations, PII, language, tone, deliverables)
  -> Playwright smoke checks     (signup, email, mobile CTA, export)
  -> load/SLO probe              (asyncio + Locust, p50/p95/error)
  -> chaos-lite probe            (429 / timeout / empty retrieval)
  -> verdict engine              (rule-based, not LLM-judged)
  -> exit code + rich TUI + cmux alert + report.json + report.html
  -> SHIP / HOLD / BLOCK
```

CLI가 source of truth. TUI는 live release-board view. HTML report는 evidence package.
cmux alert는 발표 증폭 장치이지 제품 본체가 아니다.

---

## 포지셔닝

| 도구 분류 | 답하는 질문 |
|---|---|
| 전통적 CI | 코드 테스트가 통과하는가? |
| Observability | 배포 후 무슨 일이 일어났는가? |
| LLM Gateway | 모델 요청을 어떻게 라우팅할 것인가? |
| Load Testing | 트래픽 아래 endpoint가 어떻게 동작하는가? |
| Generic QA Tool | 제품이 일반 체크리스트를 만족하는가? |
| **Agent Sail** | **이 에이전트가 이 고객의 출시 기준을 출시 전에 만족하는가?** |

> Codex는 요청받으면 코드를 고친다. Agent Sail은 출시 기준을 반복 가능하고 측정 가능하고
> 강제 가능하고 artifact로 남는 게이트로 만든다.

---

## Install

웹에서 한 줄 설치 (퍼블릭 repo일 때):

```bash
curl -fsSL https://raw.githubusercontent.com/LHANTAEK/cmux-260426-hackathon/mvp/scripts/install.sh | bash
```

릴리스 바이너리는 `v*` 태그 push 시 자동 발행:

```bash
git tag v0.1.3 && git push origin v0.1.3
```

로컬 개발 빌드:

```bash
./scripts/install-local.sh
# AGENTSAIL_INSTALL_DIR=/tmp/bin ./scripts/install-local.sh
# AGENTSAIL_INSTALL_CODEX_SKILL=0 ./scripts/install-local.sh
```

설치 스크립트는 `agentsail` 바이너리뿐 아니라 Codex 스킬(`$CODEX_HOME/skills/agentsail`) 과
`agentsail-marketplace` Codex 플러그인 등록도 같이 갱신한다.

---

## Project Init

프로젝트당 1회:

```bash
agentsail init
```

설치되는 것:

- `.claude-plugin/`, `.claude/` — Claude Code 명령, 스킬, 에이전트, 훅, 룰
- `.codex-plugin/`, `commands/`, `skills/` — Codex 플러그인 슬래시 명령
- `.agents/skills/agentsail/` — Codex repo-scope 스킬 디스커버리
- `fixtures/agentsail/` — 결정론적 데모 fixture (Slack/Notion/Gmail/GitHub export 흉내)
- `agentsail.loadtest.yaml`, `locust/agentsail/` — YAML 기반 Locust 부하테스트
- `.agentsail/` — evidence/cache/contract 디렉토리 트리
- `AGENTS.md` Agent Sail 블록

---

## Demo (데모 3개)

```bash
agentsail ci --customer finbank   --target mock:support_agent_v12 --report --soft-exit
agentsail ci --customer retailco  --target mock:support_agent_v12 --report
agentsail ci --customer acme-bank --target mock:support_agent_v12 --report --cmux-alert --soft-exit
```

기대 verdict:

```text
finbank   -> HOLD     (citation 누락 / 50명 ramp 중 31명에서 p95 5s 초과)
retailco  -> SHIP
acme-bank -> BLOCK    (missing CSV export, beta badge exposed, tone drift)
```

가장 강한 시연 장면은 `acme-bank` 실행 → cmux 사이드바 빨간 알림 점멸 →
`report.html` 자동 오픈 → `VERDICT: BLOCK` 한 컷.

---

## Slash Commands

`/agentsail:*` 네임스페이스 한 군데로 통합되어 있다 (구 `agentsail-ci` 같은 하이픈 형은 제거).

| Command | 역할 |
|---|---|
| `/agentsail:init` | 프로젝트에 Agent Sail 설치 |
| `/agentsail:collect <customer>` | Slack / Gmail / Notion / GitHub MCP에서 고객 맥락 수집 → `.agentsail/cache/<customer>/*.json` |
| `/agentsail:compile --customer <c>` | 캐시를 `customer_contract.json`으로 컴파일 |
| `/agentsail:check --customer <c> --target <t>` | 결정론적 기준 검사 + chaos-lite probe |
| `/agentsail:verdict --customer <c>` | 최신 run을 `SHIP`/`HOLD`/`BLOCK`으로 판정 |
| `/agentsail:ci --customer <c> --target <t> [--report] [--open]` | compile → check → verdict → report 일괄 |
| `/agentsail:report <run-json> [--open]` | run JSON에서 standalone HTML evidence 생성 |
| `/agentsail:loadtest <init\|tui\|run\|explain>` | 부하테스트 (아래) |
| `/agentsail:doctor` | CLI/런타임 상태 확인 |
| `/agentsail:version` | 버전 |

각 슬래시 커맨드는 `.claude/commands/agentsail/*.md` 한 곳에서 단일 정의된다 (스킬 디렉토리에는 dispatcher `agentsail` 하나만 남김).

---

## Customer Contract 모델

YAML은 editable view / fixture format이다. 제품 artifact는 컴파일된 `customer_contract.json`.

HTTP target 예시:

```yaml
customer: retailco
target:
  type: http
  endpoint: http://localhost:8000/chat
  method: POST
  prompt_field: message
  answer_field: answer
criteria:
  quality: { language: ko }
  reliability:
    expected_concurrency: 20
    max_p95_latency_ms: 5000
    max_error_rate: 0.01
```

LangGraph target 예시:

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
  quality: { citations_required: true, language: en }
  reliability:
    expected_concurrency: 50
    max_p95_latency_ms: 5000
    max_error_rate: 0.01
    fallback_on_429: short_answer
  safety:
    pii_exposure: deny
    external_email_requires_approval: true
```

---

## Target Adapters

| Type | 동작 |
|---|---|
| `langgraph` | 로컬 LangGraph app/runnable을 import해서 직접 invoke |
| `http` | 실행 중인 agent endpoint를 `httpx`로 호출 |
| `mock` | 발표 안정성을 위한 deterministic demo target |

Adapter contract:

```text
input  : prompt, scenario metadata
output : answer, citations, actions, latency, error
```

harness가 structured evidence를 반환하면 verdict engine이 `SHIP`/`HOLD`/`BLOCK`으로 변환한다.

---

## MVP Checks

### Criteria Check (결정론적)
- 필요할 때 citation 존재 여부
- 응답 언어 일치
- 외부 action에 approval metadata
- policy가 금지한 PII 노출 없음
- required scenario output 존재
- 합의 문구 / 출시 제약과의 drift 없음

### Playwright Smoke Check
- signup / auth path 완료
- 가입 후 welcome email / notification 발송
- mobile CTA, 주요 customer-facing 화면
- required export / customer deliverable 노출

### Load / SLO Probe (production capacity가 아니라 launch-readiness probe)
- `asyncio` 또는 Locust 기반 짧은 ramp
- p50, p95, error rate, 완료 사용자 수
- 실패 예: `[FAIL] load: reached only 31/50 users before p95 exceeded 5000ms`

### Chaos-lite Probe
- provider `429`
- tool timeout
- empty retrieval result

### Verdict
- `SHIP` / `SHIP WITH LIMITS` / `HOLD` / `BLOCK`
- `PATCH`는 별도 state가 아니라 `HOLD`/`BLOCK` 내부의 action lane (patch sketch, failing test, next command)

---

## Load Test

### YAML + Locust 기본 흐름

```bash
agentsail loadtest init    --config agentsail.loadtest.yaml
agentsail loadtest explain
agentsail loadtest doctor  --config agentsail.loadtest.yaml
agentsail loadtest install --config agentsail.loadtest.yaml
agentsail loadtest run     --config agentsail.loadtest.yaml --dry-run
agentsail loadtest tui     --config agentsail.loadtest.yaml --dry-run
agentsail loadtest tui     --config agentsail.loadtest.yaml
agentsail loadtest run     --config agentsail.loadtest.yaml
```

`run` / `tui`는 `.agentsail/loadtests/.venv` 를 `uv` 로 자동 생성하고 `locust` + `httpx` 를 설치한다.
`--no-install` 은 CI 이미지가 자체 Locust를 가질 때.

### 추적 메트릭 (참조: `llm-apps-monitoring-0424`)

| Metric | 단위 | SLO |
|---|---|---|
| `ttft_seconds` | seconds | p95 < 1.5s |
| `inter_token_latency_seconds` | seconds | p95 < 0.08s |
| `total_response_seconds` | seconds | p95 < 10s |
| `llm_errors_total / llm_requests_total` | ratio | < 1% |
| `request_queue_depth`, `concurrent_llm_calls`, `concurrent_sessions` | gauge | leading saturation |
| `container_memory_working_set_bytes` | bytes (display: GB) | < 80% of `resources.memory.limit` |

`resources.memory` 는 Docker Compose 스타일 (`1g`, `512m`, `alert_at: 80%`) 로 쓰면 Agent Sail이 내부에서 변환한다.

### Target 서버 요구사항

부하테스트 대상 서버는 두 endpoint만 노출하면 된다 — Agent Sail은 LLM을 직접 호출하지 않는다.

- `POST /chat` — 입력 `{"messages":[{"role":"user","content":"..."}]}`, 응답은 SSE (`data: <token>\n\n` ... `data: [DONE]`)
- `GET /metrics` — Prometheus exposition (위 메트릭 노출)

레퍼런스 구현 (LangGraph + Azure OpenAI + FastAPI + `prometheus_client`) 은 `examples/loadtest-target/` 에 있다 (있을 때).
필요한 환경변수만 정리:

```bash
AZURE_AI_FOUNDRY_ENDPOINT
AZURE_AI_FOUNDRY_API_KEY
AZURE_AI_FOUNDRY_DEPLOYMENT
AZURE_AI_FOUNDRY_API_VERSION   # 기본 2024-10-21

LANGSMITH_ENDPOINT             # 선택
LANGSMITH_API_KEY              # 선택, LANGSMITH_TRACING=false 로 두면 키 불필요
LANGSMITH_TRACING              # 부하테스트 중에는 false 권장

MAX_CONCURRENCY                # FastAPI 안의 asyncio.Semaphore 슬롯 수
```

---

## Live Release Board (rich.Live TUI)

설계 원본 [`docs/02-agent-sail/proposal.ko.md` §7](docs/02-agent-sail/proposal.ko.md) 의 ASCII 그대로:

```text
┌ Customers ───────┐ ┌ Criteria / Evidence ─────────────────────┐ ┌ Verdict ───────┐
│ acme-bank        │ │ ttft_p95_seconds: <= 1.5s                 │ │ BLOCK           │
│                  │ │ inter_token_latency_p95_seconds: <= 0.08s │ │                 │
│ Target           │ │ total_response_p95_seconds: <= 10s        │ │ Failed:         │
│ http://...:8000  │ │ error_rate: <= 1%                         │ │ 1. ttft p95     │
│                  │ │ memory_working_set: <= 80% of limit       │ │ 2. error rate   │
│ Profile          │ │ source: agentsail.loadtest.yaml           │ │                 │
│ 4 users / 30s    │ │                                            │ │                 │
└──────────────────┘ └───────────────────────────────────────────┘ └─────────────────┘

┌ Live Load Probe ───────────────────────────────────────────────────────────────┐
│ Phase: STEADY      Users: 4 / 4         RPS: 1.5                               │
│ p50: 1.20s         p95: 2.30s           error: 0.00%                            │
│                                                                                 │
│ Load   ▁▂▃▅▆▇█▇▆▆                                                               │
│ p95    ▁▁▂▃▄▅▆▇▇▇                                                               │
└────────────────────────────────────────────────────────────────────────────────┘
```

데이터 소스:
- `/metrics` 1초 폴링 → `histogram_quantile` 알고리즘으로 p50/p95 추정 (Prometheus와 동일 보간)
- `concurrent_sessions` gauge → Users
- `llm_requests_total` 카운터 차분 → RPS, sparkline
- Locust subprocess → Phase 전환만 (`RAMPING` → `STEADY` → `STOPPING` → `DONE`)

안정성을 위해 `rich.Live` 로 구현. Textual은 시간 남을 때만.

---

## cmux Side-Pane Live Demo

`cmux` 가 깔려 있으면 부하테스트 TUI를 옆 패인에 분리해서 띄울 수 있다.
사용자는 Claude Code/Codex 대화창과 라이브 보드를 동시에 본다.

```bash
SURFACE=$(cmux --json new-split right | jq -r '.surface_ref')

cmux send --surface "$SURFACE" \
  "agentsail loadtest tui --config agentsail.loadtest.yaml"
cmux send-key --surface "$SURFACE" Return
```

CLI 어휘:
- `cmux --json new-split right|left|up|down` — 신규 surface 생성, ref 반환
- `cmux send --surface <ref> "<text>"` — stdin 입력
- `cmux send-key --surface <ref> Return` — 키 입력 (TTY)
- `cmux notify` / OSC 9 escape — 사이드바 알림 점멸 (verdict 변화 신호)

해커톤 시연에서는 `acme-bank` BLOCK 직전 cmux 4개 워크스페이스 사이드바가 동시에
빨간 점멸 → `report.html` 자동 오픈으로 마무리한다.

---

## HTML Evidence Report

데모에서 가장 강한 장면은 자동으로 열린 `report.html` 의 top-line verdict.

권장 구조:
1. `VERDICT` — `SHIP` / `HOLD` / `BLOCK`
2. `WHY` — 실패한 고객 기준과 측정 evidence
3. `FIX NOW` — patch 제안, 실패 scenario, 다음 action
4. `EVIDENCE` — criteria source, probe output, load metrics, chaos-lite, screenshot/log
5. `PATCH` — 고칠 수 있을 때 generated patch sketch / failing Playwright test / concrete next command

상단 예시:

```text
VERDICT: HOLD
Customer: FinBank
Target:   langgraph:./examples/support_graph.py:app

Why held:
- 규제 답변 scenario 8개 중 3개에서 citation 누락.
- load probe가 50명 목표 전에 p95 5000ms를 초과했고 31명까지만 통과.
- provider 429에서 short fallback answer 대신 raw provider error 반환.
```

---

## 산출물

```text
.agentsail/cache/<customer>/{slack,gmail,notion,github}.json
.agentsail/contracts/<customer>-contract.json
.agentsail/runs/<customer>-run-<id>.json
.agentsail/reports/<customer>-run-<id>.html
.agentsail/loadtests/<run>/{stats_*.csv, report.html}
```

CLI exit code, `report.json`, machine-readable artifact가 release gate의 1차 출력.
TUI / cmux alert / HTML은 같은 verdict의 다른 채널이다.

---

## 리스크와 방어

| Risk | Defense |
|---|---|
| "그냥 QA/CI 아닌가?" | 전통적 CI는 generic code behavior, Agent Sail은 customer-specific launch criteria를 agent output 기준으로 검사. |
| "Codex가 하면 되는 것 아닌가?" | Codex는 코드를 고치는 도구. Agent Sail은 기준을 반복 가능, 측정 가능, 강제 가능, artifact-backed로 만든다. |
| "Load 결과가 production-accurate한가?" | capacity certification이 아니라 launch-readiness probe다. 명백한 SLO blocker만 잡는다. |
| "Verdict가 주관적인가?" | LLM judge가 아니라 customer criteria + measured evidence 기반의 rule-based 판정. |
| "왜 web dashboard가 아닌가?" | 제품은 release gate. CLI/TUI가 source of truth, HTML은 shareable evidence, cmux는 amplifier. |
| "범용 agent platform인가?" | 아니다. customer launch readiness 한 점에 집중한 release verification harness. |
| "Criteria를 손으로 입력하는가?" | 아니다. 메신저 / 메일 / Notion / GitHub / PR / staging / test에서 자동 수집 후 contract로 컴파일. |

---

## Roadmap

- 프로덕션 connector: Slack, Teams, Gmail, Notion, Linear, GitHub, PR comment
- citation 붙은 human-confirmed customer criteria
- GitHub Actions blocking check run
- 과거 verdict 비교 / regression view
- 실패 scenario replay
- Account-team web dashboard
- 좁은 criteria check용 A2A specialist evaluator

---

## 세션 메모리 (이후 편집에서도 유지)

- Agent Sail은 **고객 맥락을 아는 release gate**. 범용 orchestration platform이 아니다.
- Source: 메신저, 메일, Notion, GitHub, PR, staging, test, assistant log.
- Compiled artifact: `customer_contract.json`.
- 흐름: `collected context -> release contract -> measured evidence -> verdict`.
- CLI = exit code, `report.json`, machine-readable artifact 책임.
- TUI = live release-board view (`rich.Live`).
- cmux = visual amplifier, 제품 본체 아님.
- HTML report = shareable evidence package, 가장 강한 데모 artifact.
- MVP는 manual entry가 아니라 ingest용 local fixture 사용.
- 핵심 대비: `finbank -> HOLD`, `retailco -> SHIP`, `acme-bank -> BLOCK`.
- 가장 강한 문장: **Same agent. Different customer. Different launch gate.**

---

## License

MIT.
