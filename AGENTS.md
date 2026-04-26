# AGENTS.md

이 저장소에서 해커톤 구현을 이어받는 agent는 아래 기준을 따른다.

## 이번 해커톤 구현 대상

이번 CMUX 해커톤에서 구현할 것은 **Agent Sail — 고객별 에이전트 출시 게이트**다.

참조 문서:

- 영어 제안서: `docs/02-agent-sail/proposal.md`
- 한글 제안서: `docs/02-agent-sail/proposal.ko.md`

구현할 제품의 한 문장 정의:

> Agent Sail은 메신저, 메일, Notion, GitHub, PR, staging, test, assistant log에 흩어진 고객 맥락을 자동 수집해 `customer_contract.json`으로 컴파일하고, 에이전트가 그 고객의 출시 기준을 만족하는지 증거 기반으로 `SHIP`, `HOLD`, `BLOCK` 판정하는 release gate다.

## 반드시 유지할 포지셔닝

- 카테고리는 **Developer Tooling**이다.
- 범용 multi-agent platform이 아니다.
- full SaaS dashboard가 아니다.
- provider lock-in 제품이 아니다.
- "모든 assistant를 완벽 지원"한다고 과장하지 않는다.
- 제품 본체는 **context collectors + contract compiler + check runner + verdict engine + evidence report**다.
- cmux는 시각적 증폭 장치이고, HTML report는 증빙 산출물이다.

## MVP 범위

실제 SaaS connector를 다 붙이지 않는다. 대신 실제 integration과 같은 interface를 가진 local fixture ingest로 시연한다.

필수 구현:

- `collectors/`: messenger/email/Notion/GitHub/PR/staging/test fixture ingest
- `compiler/`: 수집된 맥락을 `customer_contract.json`으로 컴파일
- `adapters/`: 최소 `mock`, 가능하면 `http` 또는 `langgraph`
- `checks/`: criteria check, Playwright smoke, load/SLO probe, chaos-lite probe
- `engine/`: risk scoring과 `SHIP` / `HOLD` / `BLOCK` 판정
- `renderers/`: terminal summary, cmux alert, `report.json`, standalone `report.html`
- `--open`: 생성된 HTML report 자동 오픈

권장 CLI 형태:

```bash
agentsail ci --customer finbank --target langgraph:./examples/support_graph.py:app --report --open
agentsail report .agentsail/runs/finbank-run-001.json
```

생성 artifact:

```text
.agentsail/contracts/<customer>-contract.json
.agentsail/runs/<customer>-run-001.json
.agentsail/reports/<customer>-run-001.html
```

## 데모 고정 스토리

Primary demo:

- 같은 agent를 `FinBank`와 `RetailCo`에 대해 실행한다.
- `FinBank -> HOLD`
- `RetailCo -> SHIP`
- 핵심 문장: **Same agent. Different customer. Different launch gate.**

README3에서 가져온 alternate demo:

- 고객사: `acme-bank`
- 실패: `missing CSV export`, `beta badge exposed`, `tone drift`
- 장면: CLI 실행 -> cmux alert -> `report.html` 자동 오픈 -> `BLOCKED`

## 구현 우선순위

1. Local fixture 기반 context collection이 먼저다. 수동 YAML 입력 제품처럼 만들지 않는다.
2. `customer_contract.json`을 눈으로 확인 가능한 artifact로 만든다.
3. Verdict는 rule-based로 낸다. LLM judge는 omission/tone 같은 좁은 보조 check에만 쓴다.
4. HTML evidence report를 가장 먼저 polish한다.
5. TUI는 `rich.Live` 기반 release board 정도로 충분하다.
6. GitHub/CI 실제 integration은 없어도 된다. 대신 merge blocked/check failed 출력은 데모에 포함한다.

## 피해야 할 것

- full web app
- 실제 Slack/Notion/Gmail/GitHub OAuth integration에 시간 소모
- production-grade load testing
- A2A 전체 구현
- LangGraph-only 제품처럼 보이게 만들기
- cmux 시각효과가 제품 본체처럼 보이게 만들기

## 문서 수정 시 기준

Agent Sail 관련 문서를 수정할 때는 반드시 아래 문서를 먼저 확인한다.

1. `docs/02-agent-sail/proposal.ko.md`
2. `docs/02-agent-sail/proposal.md`
3. `README3.md`

특히 `README3.md`의 핵심인 **자동 context collection, release contract compilation, HTML evidence report, cmux alert, merge blocking, verdict engine**을 빠뜨리지 않는다.
