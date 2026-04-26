# 03. shiplock — "고객이 버그를 발견하기 전에, 출고가 먼저 멈춘다"

## 1. 풀려는 문제 정의

**"작은 AI 팀은 QA/QC 인력이 없고 성공 기준은 Slack·Notion·GitHub·에이전트 대화에 흩어져 있어서, LLM 산출물과 웹서비스 버그가 같은 기준으로 검수되지 못한 채 고객에게 나간다."**

근거:
- 2025 Stack Overflow Developer Survey: 개발자의 84%가 AI 도구를 사용하거나 사용할 예정이지만, 정확도를 신뢰한다는 응답은 33%뿐이고 46%는 오히려 불신
- 같은 설문: 가장 큰 불만은 **"거의 맞지만 완전히 맞지는 않은 결과물" 66%**, **"AI가 만든 코드를 디버깅하는 데 시간이 더 든다" 45.2%**
- 2026.03.10 DORA: AI 도입은 throughput 과 instability 를 같이 키우며, 작성 시간 절감분이 auditing/verification overhead 로 다시 소모됨
- 2025 GitClear: 2억1100만 changed lines 분석에서 clone code 는 증가하고 refactoring 은 감소
- 2025 Sonar State of Code: 백만 줄당 약 2,100개의 reliability issue 가 발견되며, 이런 버그는 테스트만으로는 놓치기 쉬움

## 2. 왜 지금 우리가 풀어야 하는가

이제 문제는 "AI가 코드를 쓸 수 있나?"가 아니다. **"AI가 빨리 만든 결과물을 누가 마지막에 멈추나?"** 다.

특히 소규모 AI 에이전시와 병렬 프로젝트 조직은 사람을 더 뽑아 QA 를 붙일 수 없다. 그런데 AI 덕분에 산출물 속도는 이미 빨라졌다. 즉, **출고 속도는 올라갔는데 출고 기준은 오히려 더 흩어졌다.**

해커톤 무대에서도 이 문제는 설명이 쉽다. PR, staging, 요구사항 문맥을 한 화면에 두고 "사람이 안 봤으면 그대로 고객에게 나갔을 버그"를 1회 시연으로 막아내면 된다. 화려한 에이전트 orchestration 보다 **"이 팀은 검수 인력이 없어도 출고를 멈출 수 있다"** 가 훨씬 바로 와닿는다.

## 3. 핵심 메커니즘 (3줄)

1. Slack·Notion·GitHub·이슈·PRD 에서 프로젝트 문맥을 수집해 **release checklist** 를 자동 생성
2. PR 과 staging URL 을 받아 **요구사항 불일치 / 핵심 사용자 플로우 버그 / 기존 합의와의 drift** 를 동시에 검사
3. 위험 점수가 임계치 아래면 **출고 차단** + 실패 이유 3개 + 수정 패치/테스트 초안 자동 제시

## 4. 비유로 5분 이해

> 주방장이 없는 배달 식당이 있다. 주문은 카카오톡, 메모장, 전화, 단골 손님 구두 요청으로 제각각 들어온다. 요리사는 빠르게 만들지만, 마지막에 "이게 진짜 오늘 주문한 메뉴가 맞나?"를 보는 사람이 없다. 그래서 음식은 빨리 나가지만, 반찬 하나 빠지고, 알레르기 재료가 들어가고, 국물이 새서 컴플레인이 온다.

shiplock 은 그 식당의 **마지막 출고 셰프**다. 흩어진 주문을 다시 읽어 오늘의 정답지를 만들고, 포장 직전에 음식 하나씩 대조해서 **틀리면 손님에게 보내기 전에 멈춘다.**

기술 매핑:
- 흩어진 주문 = Slack / Notion / GitHub / 에이전트 대화
- 오늘의 정답지 = release checklist
- 포장 직전 대조 = staging + PR + 테스트 동시 검사
- 출고 셰프 = shiplock

## 5. 라이브 3분 시연 시나리오

**무대 세팅**: cmux 워크스페이스 4개 동시 띄움 — `context` / `pull-request` / `staging-app` / `shiplock`. 데모용 프로젝트는 "회원가입 후 온보딩 메일 발송" 기능. 실제 고객 요구는 Notion 과 Slack 에 흩어져 있고, PR 에는 AI가 만든 버그가 섞여 있다.

| 초 | 화면 | 청중이 보는 것 | 의미 |
|---|---|---|---|
| 0s | 발표자 명령 입력 | `$ shiplock judge --project demo --pr 42 --staging http://localhost:3000` | 한 줄로 무인 출고 심사 시작 |
| 2s | `context` 워크스페이스 | Notion 요구사항 + Slack 합의 + GitHub 이슈를 읽어 checklist 6개 자동 생성 | 정답지가 원래 없었음을 먼저 보여줌 |
| 5s | `staging-app` 워크스페이스 | 회원가입 성공처럼 보이지만 실제로는 welcome email 미발송 | 사람이 놓치기 쉬운 "거의 맞는" 상태 |
| 8s | `shiplock` 워크스페이스 | `FAIL: onboarding_email_not_sent`, `FAIL: copy mismatch`, `WARN: missing mobile validation` | bug + requirement miss 를 동시에 검출 |
| 12s | `pull-request` 워크스페이스 | merge blocked 배지 + failing Playwright 테스트 생성 | 사람이 안 봐도 출고가 멈춤 |
| 18s | `shiplock` 워크스페이스 | 패치 diff 초안 + 테스트 초안 자동 제시 | 찾기만 하는 툴이 아니라 바로 고칠 수 있음 |
| 25s | 발표자 한 문장 | "QA 팀이 없어도, 고객이 보기 전에 출고가 스스로 멈춥니다." | 헤드라인 |

## 6. 9시간 빌드 분해

| 시간 | 마일스톤 |
|---|---|
| 08:00–08:30 | 환경 셋업 (`uv`, `playwright`, Gemini SDK, SQLite, cmux 동작 확인) |
| 08:30–10:00 | 문맥 수집기 구현 (Slack/Notion/GitHub 대신 데모용 로컬 fixtures ingest) |
| 10:00–11:30 | checklist 생성기 + risk scoring 로직 |
| 11:30–12:30 | staging 플로우 검사 (핵심 사용자 여정 2~3개 Playwright 시나리오) |
| 12:30–13:00 | 점심 + 데모용 버그 시드 심기 |
| 13:00–14:30 | PR 판정 / merge block 출력 / failing test 생성 |
| 14:30–15:30 | 패치 초안 생성 + rich 기반 터미널 diff 출력 |
| 15:30–16:30 | cmux 4워크스페이스 시연 시퀀스 연결 |
| 16:30–17:00 | 카피 polish + 발표용 한 줄 헤드라인 정리 |
| 17:00–18:00 | 최종 리허설 + fallback 루트 점검 |

**라이브러리**: `google-genai` 또는 `vertexai`, `playwright`, `sqlite3`, `rich`, `fastapi`(선택)

## 7. 리스크 + 대응 + 채점

**3축 채점**: ① 한 컷 시연 9/10 / ② 헤드라인 9/10 / ③ 심사관 직결 8/10 — **합 26/30**

**리스크**:
- (중) 외부 SaaS 실제 연동까지 가면 시간이 터질 수 있음 → 해커톤 빌드는 로컬 fixture + GitHub PR mock 으로 제한
- (중) "체크리스트 생성"이 추상적으로 보일 수 있음 → 데모에서 checklist 를 먼저 보여주고 바로 failing flow 로 연결
- (낮음) 패치 초안 품질이 애매할 수 있음 → patch 제안이 약하면 failing test + block 결과만으로도 메시지는 성립

**Fallback**: PR 연동이 늦어지면 `shiplock judge` 단일 CLI 로 다운그레이드하고, `PASS / FAIL / PATCH` 3단 출력만 남긴다. 본체는 여전히 "출고 전 자동 판정"이다.

**가장 큰 강점**: AI가 만든 결과물의 속도를 자랑하지 않고, 그 속도 때문에 생긴 **검수 공백**을 정면으로 친다. 문제 정의가 선명하고, 데모가 쇼가 아니라 **"원래 나갔을 버그가 여기서 멈췄다"** 는 증거로 끝난다.
