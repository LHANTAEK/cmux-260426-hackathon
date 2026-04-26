# 03. ShipCheck — "Agent CI before you launch"

## 1. 풀려는 문제 정의

**"에이전트는 데모에서 성공해도 배포에서 실패한다. 코드에는 CI가 있는데, 에이전트에는 CI가 없다."**

근거:
- 기업 88%가 에이전트 보안 사고를 경험했으나, 14.4%만 정식 보안 승인 후 배포
- EU AI Act 고위험 의무 2026-08-02 시행 → 배포 전 검증 증빙 수요 급증
- 기존 도구는 한 축만 커버:
  - Flakestorm: 프롬프트 레벨 chaos만
  - Patronus AI: 출력 품질 평가만
  - Arize/Langfuse/AgentOps: 배포 *후* 관찰만
  - GuideLLM: LLM 서버 벤치마크만
  - Cisco AI Scanner: MCP 정적 스캔만
- **load + policy + chaos 세 축을 합쳐서 SAFE/HOLD/BLOCK 판정을 내리는 도구는 없다**

## 2. 포지셔닝

**카테고리**: Agent CI
**제품명**: ShipCheck

> "Software has CI before deployment. Agents need CI before launch."
> "코드는 배포 전에 CI를 돌립니다. 에이전트도 사람들에게 열기 전에 CI가 필요합니다."

기존 제품과의 차별화:

| 도구 | 역할 | 비유 |
|---|---|---|
| LiteLLM | 모델 호출 게이트웨이 | runtime gateway |
| Langfuse / LangSmith | 관측·평가 플랫폼 | observability |
| **ShipCheck** | **배포 전 CI 게이트** | **pre-ship test gate** |

핵심 차별점:
> "LiteLLM routes model requests. We orchestrate specialist agents to decide whether an entire agent workflow is safe to launch."

## 3. 아키텍처

### 3.1 LangGraph Supervisor Graph

메인 orchestrator가 전문 에이전트들에게 검사를 분배하고, 결과를 모아 판정을 내린다.

```
                    ┌─────────────────┐
                    │  Supervisor     │
                    │  (LangGraph)    │
                    └──────┬──────────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
     ┌────────────┐ ┌───────────┐ ┌──────────┐
     │ PolicyCheck│ │ ChaosCheck│ │ LoadCheck │
     │ Agent      │ │ Agent     │ │ Agent     │
     └────────────┘ └───────────┘ └──────────┘
              │            │            │
              └────────────┼────────────┘
                           ▼
                    ┌─────────────────┐
                    │  VerdictAgent   │
                    │  (aggregate)    │
                    └──────┬──────────┘
                           │
                           ▼
                    ┌─────────────────┐
                    │  HITL Interrupt  │
                    │  (approve/reject)│
                    └─────────────────┘
```

### 3.2 A2A-style 서브 에이전트

각 에이전트는 A2A 인터페이스로 통신하며, 독립적으로 교체·확장 가능:

- **PolicyCheckAgent**: timeout 유무, max tool calls, budget cap, fallback 설정 감사
- **ChaosCheckAgent**: 429 주입, timeout 주입, malformed tool output 주입 → 생존 여부 판정
- **LoadCheckAgent**: 동시 요청 증가 → safe/degraded/fail 지점 계산

### 3.3 Human-in-the-loop

LangGraph `interrupt()` 기반:
- **SHIP**: 바로 진행
- **SHIP WITH LIMITS**: 제한 조건 표시 + 승인 요청
- **HOLD**: 수정 필요 항목 제시 + 거절 (수정 후 재실행)
- **BLOCK**: 심각한 결함 → 기본 거절

## 4. MVP 테스트 항목 (4개)

### 4.1 Policy Test
- timeout 설정 있는지
- max tool calls 제한 있는지
- budget cap 있는지
- fallback 경로 있는지

### 4.2 Chaos Test
- 429 (rate limit) 주입 시 graceful degradation 여부
- timeout 주입 시 적절한 에러 핸들링 여부
- malformed tool output 주입 시 크래시 여부

### 4.3 Load Test
- 동시 요청 증가시키며 safe concurrency 측정
- degraded 구간 시작점 측정
- fail 지점 측정

### 4.4 Verdict Card
- SHIP / SHIP WITH LIMITS / HOLD / BLOCK 판정
- 근거 3줄 요약
- 수정 필요 항목 리스트

## 5. CLI 인터페이스

### 명령어

```bash
shipcheck init              # agent.yaml 템플릿 생성
shipcheck run ./agent.yaml  # 전체 CI 실행
shipcheck report            # 마지막 결과 리포트
shipcheck ci ./agent.yaml   # GitHub Actions 스타일 CI 모드
```

### 출력 예시

```
Agent CI: demo-agent

[PASS] policy: tool timeouts configured
[PASS] safety: no unbounded tool loop detected
[WARN] load: degraded after 38 concurrent users
[FAIL] chaos: provider 429 has no fallback

VERDICT: HOLD

Required before launch:
1. Add fallback for provider 429
2. Set max concurrency to 25
3. Add 3s timeout to search_tool
```

## 6. 구현 모듈

```
shipcheck/
├── cli.py                   # CLI 엔트리포인트 (typer/click)
├── supervisor_graph.py      # LangGraph state + nodes
├── a2a_clients.py           # A2A-style agent 호출 wrapper
├── agents/
│   ├── policy_checker.py    # 정책 감사 에이전트
│   ├── chaos_checker.py     # 장애 주입 에이전트
│   └── load_checker.py      # 부하 테스트 에이전트
├── verdict.py               # 판정 로직 + 집계
├── risk_rules.py            # unsafe 조건 룰 정의
├── report_renderer.py       # CLI verdict 출력 (rich)
├── approval.py              # HITL interrupt/resume handling
└── schemas/
    └── agent_spec.py        # agent.yaml 스키마 정의
```

## 7. 데모 흐름 (3분)

| 순서 | 발표자 액션 | 화면 | 의미 |
|---|---|---|---|
| 1 | 한 줄 소개 | 슬라이드 | "에이전트는 데모에서 성공해도 배포에서 실패합니다" |
| 2 | 문제 제기 | 슬라이드 | "코드는 CI가 있는데, 에이전트에는 CI가 없습니다" |
| 3 | 명령 실행 | `shipcheck ci ./agent.yaml` | 라이브 데모 시작 |
| 4 | 에이전트 실행 | LangGraph supervisor 시작, 3개 에이전트 병렬 호출 | multi-agent CI 동작 |
| 5 | 결과 반환 | policy PASS, load WARN, chaos FAIL | 각 축 결과 |
| 6 | 판정 | `VERDICT: HOLD` + 수정 항목 | 배포 불가 판정 |
| 7 | 수정 후 재실행 | fallback 추가 → `shipcheck ci` 재실행 | 수정 반영 |
| 8 | 재판정 | `VERDICT: SHIP WITH LIMITS` | 조건부 배포 가능 |
| 9 | HITL | `Ship with limits? [approve/reject]` → approve | 사람이 최종 승인 |

## 8. 인프라 요구사항

**해커톤 데모**: MacBook 하나로 충분
- LangGraph supervisor + 3 에이전트 = Python 프로세스 (256MB ~ 1GB RAM)
- LLM은 클라우드 API 호출
- 부하 테스트는 asyncio/httpx → 4 vCPU / 8GB RAM이면 넉넉

**프로덕션**:
- 단일 노드: 4 vCPU / 8-16GB RAM
- 클라우드 비용: $50-150/월 + LLM API 비용 (변동)

## 9. 12시간 빌드 분해

| 시간 | 마일스톤 |
|---|---|
| 0:00–0:30 | 환경 셋업 (uv, LangGraph, A2A SDK, rich) |
| 0:30–2:00 | agent.yaml 스키마 + PolicyCheckAgent |
| 2:00–3:30 | ChaosCheckAgent (429/timeout/malformed 주입) |
| 3:30–5:00 | LoadCheckAgent (asyncio 동시 요청) |
| 5:00–6:00 | LangGraph supervisor graph 연결 |
| 6:00–7:00 | VerdictAgent + risk_rules.py |
| 7:00–8:00 | HITL interrupt/resume + approval.py |
| 8:00–9:00 | CLI (shipcheck init/run/ci/report) |
| 9:00–10:00 | report_renderer.py (rich 출력 polish) |
| 10:00–11:00 | 데모 시나리오 + 더미 에이전트 준비 |
| 11:00–12:00 | E2E 리허설 ×3 + 발표 준비 |

**라이브러리**: `langgraph`, `a2a-sdk`, `typer`, `rich`, `httpx`, `pyyaml`

## 10. 리스크 + 대응

| 리스크 | 확률 | 대응 |
|---|---|---|
| A2A 풀스펙 구현 시간 부족 | 높음 | 로컬 subprocess/HTTP로 구현, 인터페이스만 A2A 스타일. "A2A-ready specialist agents" 포지셔닝 |
| LLM API 지연으로 데모 느림 | 중 | 데모용 mock 응답 캐시 준비 |
| Load test가 로컬에서 의미 없어 보임 | 중 | "safe concurrency" 숫자 + degradation curve 시각화로 설득력 확보 |
| 판정이 임의적으로 보일 수 있음 | 낮음 | risk_rules.py에 명시적 룰 정의, 각 FAIL/WARN에 근거 표시 |

## 11. 확장 로드맵 (v2)

- **Replay Test**: 실패 trace 재현 → 어느 step에서 터졌는지 표시
- **GitHub Actions 연동**: CI/CD 파이프라인에 `shipcheck` 스텝 추가
- **Verdict History**: 배포 판정 이력 대시보드
- **Custom Rules**: 팀별 정책 룰 정의 (`.shipcheckrc`)
- **A2A 원격 에이전트**: 외부 전문 에이전트와 실제 A2A 프로토콜 통신
