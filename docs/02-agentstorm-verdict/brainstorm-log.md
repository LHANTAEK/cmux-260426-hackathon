# agentstorm verdict — 브레인스토밍 로그

> 출처: Claude.ai 대화 (2026-04-26 09:01~09:50)
> 원본 아이디어: LLM App 오토스케일 및 모니터링 구성 테스트 계획 (로컬 POC)

---

## 아이디어 진화 과정 (5단계)

### 1단계: LLM 모니터링 CLI (탈락)
- 원안: 로컬 compose 환경에서 LLM 워크로드 스케일링 지표 검증
- SLO: TTFT p95 < 3s, ITL p95 < 80ms, Total p95 < 10s
- **탈락 이유**: "리서치 페이퍼"에 가까움. 12시간에 5개 phase 불가능. 해커톤은 데모 중심.

### 2단계: agentctl — 에이전트 디버거 (탈락)
- "Langfuse가 웹으로 이미 함"이 치명적
- debug/explain/replay 3개 명령 구상
- **탈락 이유**: Claude한테 trace 던지고 "분석해줘" 하면 80%는 나옴 → "Claude 루프로 안 됨?" 방어 불가

### 3단계: agentstorm — 에이전트 부하 테스터 (보류)
- 핵심 전환: "에이전트 잘 만들어서 딜리버해도, 사용자가 많으면 박살이 나요"
- 에이전트 endpoint에 동시 N명 시뮬레이션, breaking point 자동 탐지
- **보류 이유**: 일회성 도구, 결론이 없음

### 4단계: agentstorm autopilot — SRE 에이전트 (보류)
- "너의 에이전트를 위한 SRE 에이전트. AI가 너의 AI를 망가뜨리고, AI가 고친다."
- LLM이 hypothesis → 실험 설계 → 결과 해석 → 패치 생성 → 검증
- **보류 이유**: 에이전틱하지만 결론이 없음. "몇 명까지 가능한지" 답이 안 나옴

### 5단계: agentstorm verdict — capacity verdict (최종안)
- **핵심 메시지**: "에이전트 만들었는데 몇 명까지 받을 수 있는지 모르세요? 2분 후에 답이 나옵니다."
- 두 모드: `verdict` (사전, 2분) + `watch` (사후, 실시간)
- 노드별 capacity 분해가 차별화 핵심

---

## 최종안 상세

### 한 줄 정의
에이전트 시스템의 노드별 capacity를 분해하고, production에서 깨지기 전에 알려주는 CLI.

### 두 가지 모드

```bash
# 사전: 2분 안에 verdict
$ agentstorm verdict ./my_agent.py
# → 노드별 capacity 분해 → "synthesizer 30명, 시스템 한계 30명"

# 사후: production 감시
$ agentstorm watch --profile ./agentstorm.profile.json
# → "지금 22명, 30명까지 8명 남음, 추세상 4분 후 도달"
```

### Verdict 카드 출력 (데모 hero)

```
═══════════════════════════════════════════════════
  VERDICT — node-level capacity breakdown
═══════════════════════════════════════════════════

  Node             Safe    Degraded   Hard Limit
  ────────────────────────────────────────────────
  planner           450       —          —
  search_tool       200      280        320
  synthesizer        30       47         62    ← bottleneck
  critic            180      210        250
  
  System capacity = min(all nodes) = 30 concurrent

  At 30 users:
    p95 latency     2.1s
    cost/request    $0.04
    monthly @ 50%   $340

═══════════════════════════════════════════════════
  SCALING ROADMAP
═══════════════════════════════════════════════════

  Want 100? → parallelize synthesizer (3 workers)
  Want 500? → + provider rate limit mitigation
  Want 1k+  → architectural rework needed
```

### Watch 카드 출력

```
  Now: 22 concurrent users

  ┌─ Capacity trajectory ──────────────────────┐
  │ ●━━━━━━━━━━━━━●━━━━━━━━○                    │
  │ now=22       warn=25   break=30 (synthesizer)│
  └─────────────────────────────────────────────┘

  Trend: +2 users/min → break in ~4 min
  
  synthesizer queue: 8 (baseline: 2)  ⚠ rising
  search_tool queue: 1 (baseline: 1)  ✓ stable
```

### 차별화 포인트
- GuideLLM: vLLM 같은 모델 서버 대상. 에이전트 그래프 구조 안 다룸.
- NVIDIA NeMo Agent Toolkit: NVIDIA 생태계 락인.
- TestSprite: 일반 웹앱 중심. 에이전트 그래프 인지 없음.
- **본 도구**: LangGraph 노드 단위로 분해 → "시스템 한계 = 가장 약한 노드"

### 경쟁 도구 (검색 결과)
- **GuideLLM** (vLLM, Red Hat/Neural Magic): LLM 배포 성능 벤치마킹. sweep 프로필로 자동 rate 탐색. "How many servers do I need?" 질문에 답. **가장 위협적.**
- **NVIDIA NeMo Agent Toolkit**: LangGraph deep-research 에이전트를 1000명까지 스케일하는 3단계 프로세스.
- **TestSprite**: 자율 부하 테스트 에이전트, self-repair 포함.
- **Gatling AI/LLMs**, **LLM Locust**, **Last9**, **Langfuse**, **Arize**, **OpenObserve** 등

---

## 기각된 후보들

### 후보 1: agentprobe (AI Safety 트랙)
- "내 LangGraph/CrewAI 에이전트가 프롬프트 인젝션·탈옥·툴 오용에 얼마나 약한지 30초 만에 점수 매겨주는 CLI"
- 차별화: 에이전트 그래프 구조를 인식하는 레드팀 (garak/PyRIT는 단일 LLM)
- AIM Intelligence 트랙과 정확히 정렬
- **기각 이유**: 사용자가 "난 내꺼 하고 싶어" — capacity verdict 선택

### 후보 X1: agent-vs-agent (결투장)
- 적대적 에이전트가 자율적으로 부하 + 엣지케이스 + 어뷰징
- "와" 강도 최고, 12시간 실현가능성 최저

### 후보 X2: agent-twin (디지털 트윈)
- production trace 학습 → 비용 0원 시뮬레이션
- 카테고리 새로움, 실증 어려움

### 후보 X3: dryrun (PR bot)
- PR마다 capacity 영향 자동 분석 + GitHub 코멘트
- Developer Tooling fit 좋음, GitHub 봇 인프라 시간 소요

---

## 핵심 인사이트 (대화에서 도출)

1. **"Claude 루프로 안 됨?" 방어**: verdict를 내려면 진짜 부하를 진짜 시스템에 쏘고 진짜 메트릭을 봐야 함. Claude는 "87명"이라는 실측 숫자를 만들어낼 수 없음.

2. **"남이 쓰는 거"가 핵심**: "Claude Code는 너 한 명이 쓰는 거고, 너의 에이전트는 1000명이 쓴다."

3. **사전이 사후의 뇌가 됨**: verdict에서 학습한 breaking point 모델을 watch가 그대로 씀. 일반 APM은 깨지는 모양을 모르니까 깨지고 나서야 알려줌.

4. **결론이 있어야 함**: 그래프가 아니라 숫자 한 개. "87명까지 안전합니다."

5. **트랙 fit**: Developer Tooling. "에이전트 오케스트레이션" 키워드 활용. 피칭 한 줄: "에이전트 시스템에는 kubectl이 없다. 우리가 만들었다."

---

## 12시간 빌드 플랜 (최종)

| 시간 | 마일스톤 |
|---|---|
| 0~1h | 데모 에이전트 (LangGraph 4노드) + mock-llm endpoint |
| 1~3h | 노드별 메트릭 수집 (LangGraph callback handler) |
| 3~5h | breaking point 탐지 + verdict 계산 |
| 5~7h | watch 모드 (profile → trajectory → ETA) |
| 7~8h | autopilot narration (의사결정 트리 + LLM narration) |
| 8~10h | verdict 카드 디자인 + 데모 리허설 |
| 10~11h | README + 슬라이드 |
| 11~12h | 버퍼 |

### 버릴 것
- Grafana / VictoriaMetrics / Prometheus
- 진짜 LLM 호출 부하 테스트 (mock 기본)
- K8s manifest 생성
- 다양한 프레임워크 (LangGraph only)
- 진짜 자율 LLM 루프

---

## 피칭 구조 (3분)

| 시간 | 내용 |
|---|---|
| 0:00-0:20 | 문제: "몇 명까지 받을 수 있나요? — 모르죠?" |
| 0:20-0:40 | 기존 도구 한계: 모델 서버 vs 에이전트 그래프 |
| 0:40-1:50 | 라이브 데모: verdict (2분 안에 끝남) |
| 1:50-2:30 | 라이브 데모: watch (trajectory → break 경고) |
| 2:30-2:50 | 차별화 + 메시지 |
| 2:50-3:00 | 마무리: "2분, 터미널 한 줄." |
