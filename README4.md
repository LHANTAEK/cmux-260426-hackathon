# 04. release-harness-on-cmux — "고객이 화내기 전에 잘못된 출고를 cmux와 HTML 리포트로 막아 세운다"

## 1. 풀려는 문제 정의

**"작은 AI 팀은 Slack·Notion·GitHub 에 흩어진 고객사 맥락을 놓친 채 웹 기능·LLM 산출물·설정값을 한 번에 출고하고, QA/QC 인력이 없어 고객이 먼저 문제를 발견한다."**

근거:
- 2026 *OX Security AppSec Benchmark Report*: 216M+ findings, 조직당 평균 865K alerts, 795 critical issues. **AI 가 개발 속도를 올릴수록 마지막 검수는 더 비어 간다**
- 2026.02 *Check Point Research*: Claude Code 프로젝트 파일 취약점 (CVE-2025-59536, CVE-2026-21852) 공개. **이제 repo 설정·도구 연결·자동화 흐름 자체가 실행 레이어**라서 눈으로 보는 QA 는 더 약해짐
- GitHub 는 `gh models eval`, OpenAI 는 trace grading / evals 를 공식화. 업계도 이미 **"AI workflow 는 릴리즈 전 재현 가능한 검증이 필요하다"** 는 쪽으로 움직이는 중
- 그런데 지금 비어 있는 건 **모델 자체의 품질 평가**가 아니라, **고객사별 맥락이 섞인 통합 deliverable 의 최종 출고 게이트**

## 2. 왜 지금 우리가 풀어야 하는가

`Developer Tooling` 카테고리에서 심사관이 바로 이해하는 문제는 단순하다. **AI 코딩은 빨라졌는데, 마지막 QA 한 사람은 아무도 없다.**

GitHub 와 OpenAI 가 이미 eval·trace·deployment protection primitives 를 깔아 줬다는 것은, 시장이 "릴리즈 전 자동 판정" 을 받아들일 준비가 됐다는 뜻이다. 그런데 그들이 아직 안 푼 구멍이 있다. **Slack·Notion·GitHub 에 흩어진 고객사 맥락을 읽고, 실제 client release 를 막아 주는 도구**는 비어 있다.

여기서 포인트는 `AI coding assistant` 에 종속된 플러그인이 아니라, **어떤 에이전트·CLI·웹 타깃에도 붙일 수 있는 headless release harness** 라는 점이다. 하네스는 입력 어댑터와 판정 엔진, 출력 채널만 바꾸면 된다. 즉 `oh-my-braincrew` 같은 오케스트레이션 하네스 위에도 얹을 수 있고, 단일 CLI 에도 붙일 수 있다.

또한 시연 무대가 cmux 다. 보통의 release gate 는 CI 로그 한 줄로 끝나지만, 우리는 **어느 컨텍스트에서 무엇이 어긋났는지**를 cmux 워크스페이스 4개에서 동시에 보여주고, 동시에 `report.html` 을 열어 비기술 심사자도 이해 가능한 증빙물을 바로 보여줄 수 있다. 즉, `개발 도구`이면서도 **심사장에서 한 컷으로 보이는 도구**가 된다.

## 3. 핵심 메커니즘 (3줄)

1. 발표자가 넘긴 `Slack thread + Notion page + GitHub PR` 를 읽어 고객사 요구사항·금지사항·deliverable 범위를 **customer contract JSON** 으로 컴파일
2. 하네스가 `report.md + config.json + staging URL` 을 동시에 검사해 하드룰 위반, 웹 회귀, LLM 산출물 맥락 누락을 **release risk** 로 판정
3. 위험 점수가 기준을 넘으면 `client-release approve` 를 즉시 차단하고, cmux OSC 9 알림 + `report.html` 생성/오픈으로 **빨강 연쇄 점멸 + 증빙 리포트** 를 동시에 띄움

## 4. 비유로 5분 이해

> 작은 케이터링 업체가 오늘 4개 고객사 행사용 도시락을 한꺼번에 내보낸다. 요구사항은 여기저기 흩어져 있다. 슬랙엔 "A사는 땅콩 금지", 노션엔 "B사는 비건", 깃엔 새 라벨 프린터 설정이 들어 있다. 팀은 적고 바쁘다 보니 도시락은 예쁘게 포장되지만, **A 고객 박스에 땅콩 소스가 들어간 채 트럭이 출발**한다. 고객이 뚜껑을 열고 나서야 사고를 안다.

release-harness 는 출고장 문 앞에 서 있는 **상차 검사원**이다. 박스가 예쁘게 포장됐는지보다, **누구 박스에 무엇이 잘못 들어갔는지**를 출발 직전에 막는다. 그리고 막은 이유를 종이 검수표처럼 바로 출력해 준다.

기술 매핑:
- 주문 메모 = Slack·Notion·GitHub 에 흩어진 고객사 컨텍스트
- 도시락 박스 = `웹 기능 + LLM 산출물 + 설정값` 이 묶인 client release bundle
- 상차 검사원 = release-harness
- 출고장 빨간 경광등 = cmux OSC 9 연쇄 점멸
- 검수표 = 자동 생성되는 `report.html`

## 5. 라이브 3분 시연 시나리오

**무대 세팅**: cmux 워크스페이스 4개 동시 띄움 — `slack-thread` / `notion-spec` / `github-pr` / `release-lane`.  
데모용 고객사 `acme-bank` 는 3개 요구사항을 가진다: `CSV export 필수`, `white-label only`, `보고서 문구는 enterprise tone`.

| 초 | 화면 | 청중이 보는 것 | 의미 |
|---|---|---|---|
| 0s | 발표자 명령 입력 | `$ ./client-release approve acme-bank --pr 142 --notion ACM-12 --slack 171223.0123 --open-report` | 출고 버튼을 CLI 한 줄로 압축 |
| 1s | `release-lane` 로그 | `collecting context...` | 단순 QA 가 아니라 컨텍스트 수집부터 시작 |
| 2s | `slack-thread` 하이라이트 | `CSV export 꼭 포함해주세요` 노란 강조 | 고객의 실제 요구사항 출처 |
| 3s | `notion-spec` 하이라이트 | `No beta badge / white-label only` 빨간 밑줄 | 문서상 금지사항 |
| 4s | `github-pr` diff | `beta_banner=true` / export route 없음 | 코드와 맥락이 충돌 |
| 5s | **cmux 사이드바 4개 빨강 연쇄 점멸** | 시각적 폭발 | "출고 중지"를 한 컷으로 전달 |
| 7s | `release-lane` 결과 | `BLOCKED: missing CSV export / beta badge exposed / tone drift in report.md` | 웹 + 문서 + 설정을 한 번에 잡음 |
| 8s | 브라우저 또는 기본 뷰어 | `report.html` 자동 오픈, failed evidence 3개 카드 | 비기술 심사자도 즉시 이해하는 증빙물 |
| 10s | 발표자 한 문장 | "QA 한 사람을 못 두는 팀도, 고객이 화내기 전에 출고를 막고 그 이유를 한 장 리포트로 남길 수 있습니다" | 헤드라인 |

## 6. 9시간 빌드 분해

| 시간 | 마일스톤 |
|---|---|
| 08:00–08:30 | 환경 셋업 (`uv`, `slack_sdk`, `notion-client`, GitHub token, cmux OSC 9 확인) |
| 08:30–09:30 | `Slack thread / Notion page / GitHub PR` fetcher 3종 |
| 09:30–11:00 | 고객사 요구사항·금지사항 추출 → `customer_contract.json` 생성 |
| 11:00–12:00 | release harness adapter (`report.md`, `config.json`, `staging URL`) |
| 12:00–12:30 | 하드룰 검사기 + 간단한 LLM judge (tone / omission 전용) |
| 12:30–13:00 | 점심 + 데모용 fixture 시드 (`acme-bank`) |
| 13:00–14:30 | `client-release approve` CLI + block/allow flow 완성 |
| 14:30–15:30 | Playwright smoke test + 실패 시 screenshot 캡처 |
| 15:30–16:15 | `report.json` → `report.html` 렌더 + `--open-report` 동작 |
| 16:15–17:00 | cmux OSC 9 연동 + 빨간 diff 시각화 |
| 17:00–17:30 | E2E 시연 리허설 ×3 |
| 17:30–18:00 | fallback 준비 + 발표 슬라이드 1장 |

**라이브러리**: `slack_sdk`, `notion-client`, `PyGithub` 또는 GitHub REST, `playwright`, `pydantic`, `rich`, `openai`, `jinja2`

## 7. 리스크 + 대응 + 채점

**3축 채점**: ① 한 컷 시연 9/10 / ② 헤드라인 9/10 / ③ 심사관 직결 8/10 — **합 26/30**

**리스크**:
- (중) Slack / Notion live API 연동이 토큰·권한·rate limit 에 걸릴 수 있음 → **한 release 당 1 thread + 1 page** 만 읽고, 직전 snapshot 캐시로 대응
- (중) LLM judge 가 흔들릴 수 있음 → 판정의 80% 는 하드룰로 두고, LLM 은 tone / omission 만 담당
- (낮음) HTML 리포트가 구현 시간을 잡아먹을 수 있음 → 초기 버전은 `jinja2` 정적 템플릿 1장으로 제한
- (낮음) staging smoke test 가 깨질 수 있음 → Playwright 실패 시에도 config / report 검사만으로 block 사유는 유지

**Fallback**: live API 가 불안하면 `fixtures/slack.json`, `fixtures/notion.json`, `fixtures/pr.diff` 로 다운그레이드. HTML 이 늦어지면 `report.json + rich summary` 로 먼저 가고, `open-report` 는 정적 HTML 1장만 붙인다. 본체는 그대로 `출고 직전 block` 이고, cmux 점멸과 diff 시연은 유지된다.

**가장 큰 강점**: 문제 정의가 현실적이다. 작은 AI 팀, 많은 고객사 컨텍스트, 부족한 QA/QC, 고객이 먼저 발견하는 사고. 그리고 해결 방식도 단순하다. **고객이 보기 전에 한 번 막아 주는 release harness**. `Developer Tooling` 으로 설명이 가장 쉽고, 필요하면 `Business & Applications` 로도 바로 번역 가능하다.

## 부록 A. 새 세션용 핵심 맥락

이 문서의 현재 1순위 해석은 `generic agent platform` 이 아니라 **customer-aware release harness** 다.

한 문장 요약:

> 여러 고객사를 동시에 상대하는 작은 AI 팀이, 흩어진 고객사 맥락을 자동 수집해 통합 deliverable 을 출고 직전에 검사하고, 위험하면 차단하며, 그 이유를 cmux 경고와 HTML 리포트로 남기는 도구.

반드시 유지할 것:
- 본체는 `판정 엔진` 이다. cmux 는 시각 효과 채널이고, HTML 은 증빙 채널이다.
- 입력은 `Slack + Notion + GitHub` 3개가 기본이다.
- 출력은 `exit code + rich TUI + cmux alert + report.json + report.html` 이다.
- 최종 액션은 `allow` 또는 `block` 이다.
- CLI 가 source of truth 다. 웹은 대시보드가 아니라 생성물이다.

바꾸지 말아야 할 메시지:
- "고객이 화내기 전에 잘못된 출고를 막는다"
- "QA 한 사람을 못 두는 작은 AI 팀의 마지막 출고 게이트"
- "cmux 는 제품 본체가 아니라 심사장 증폭기"

절대 확장하지 말 것:
- 범용 에이전트 오케스트레이션 플랫폼
- full SaaS dashboard
- provider lock-in 제품
- "모든 assistant 지원" 같은 과장

권장 제품 형태:
- `client-release approve ... --open-report`
- 실행 후 `report.json` 과 `report.html` 생성
- macOS 에서는 `open report.html`, Linux 에서는 `xdg-open report.html` 로 분기 가능

하네스 관점에서의 모듈 분해:
- `collectors/`: Slack, Notion, GitHub 컨텍스트 수집
- `compiler/`: 수집한 맥락을 `customer_contract.json` 으로 컴파일
- `checks/`: hard rules, Playwright smoke, LLM omission/tone judge
- `engine/`: score 산정, block/allow 판정
- `renderers/`: terminal summary, cmux message, html report

`oh-my-braincrew` 와의 관계:
- `oh-my-braincrew` 는 에이전트 실행/오케스트레이션 하네스의 레퍼런스다
- 여기서는 그 철학을 빌리되, 목표를 `release evaluation harness` 로 좁힌다
- 즉 "무엇이든 orchestrate" 가 아니라 "무엇이든 evaluate before release" 다

데모 고정값:
- 고객사 이름: `acme-bank`
- 실패 포인트 3개: `missing CSV export`, `beta badge exposed`, `tone drift`
- 시연 핵심 장면: cmux 빨강 점멸 직후 `report.html` 자동 오픈
