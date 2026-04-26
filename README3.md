# 03. shiplock-harness — "AI가 만든 결과물은 증거 없이 출고되지 않는다"

> **새 세션용 프레이밍 메모**
> - 이 후보는 `AI Safety`가 아니라 **`Developer Tooling` 후보**다.
> - 제품의 본체는 `범용 멀티에이전트 플랫폼`이 아니라 **release verification harness** 다.
> - `Claude/Codex/Gemini` 중 무엇으로 만들었는지가 핵심이 아니라, 그 위를 감싸는 **assistant-agnostic wrapper** 가 핵심이다.
> - `TUI/GUI` 는 옵션 채널이다. 데모의 주인공은 **standalone HTML evidence report + 자동 open** 이다.
> - 절대 `무엇이든 붙일 수 있는 하네스`처럼 말하지 않는다. 항상 **"출고를 멈추는 증거 시스템"** 으로 설명한다.

## 1. 풀려는 문제 정의

**"작은 AI 팀과 에이전시는 QA/QC 인력이 없고 성공 기준은 Slack·Notion·GitHub·에이전트 대화에 흩어져 있어서, LLM 산출물과 웹서비스 버그가 같은 기준으로 검수되지 못한 채 고객에게 나간다."**

근거:
- 2025 Stack Overflow Developer Survey: 개발자의 84%가 AI 도구를 사용하거나 사용할 예정이지만, 정확도를 신뢰한다는 응답은 33%뿐이고 46%는 오히려 불신
- 같은 설문: 가장 큰 불만은 **"거의 맞지만 완전히 맞지는 않은 결과물" 66%**, **"AI가 만든 코드를 디버깅하는 데 시간이 더 든다" 45.2%**
- 2026.03.10 DORA: AI 도입은 throughput 과 instability 를 같이 키우며, 절약된 작성 시간이 검증 비용으로 다시 소모됨
- 2025 GitClear: 2억1100만 changed lines 분석에서 clone code 는 증가하고 refactoring 은 감소
- 2025 Sonar State of Code: 백만 줄당 약 2,100개의 reliability issue 가 발견되며, 이런 버그는 테스트만으로는 놓치기 쉬움

이 문제의 본질은 "AI가 실수한다"가 아니다. **"AI가 빨리 만든 결과물을, 아무도 마지막에 출고 판정하지 못한다"** 다.

## 2. 왜 지금 우리가 풀어야 하는가

요즘 작은 팀의 병목은 구현이 아니라 **마지막 책임** 이다. Claude든 Codex든 Gemini든, 이미 코드는 빨리 나온다. 하지만 고객에게 나가기 직전 누가 **"이 결과물은 증거 없이 못 나간다"** 고 말해주는가가 비어 있다.

그래서 이 후보는 QA 자동화 SaaS 가 아니라, `oh-my-braincrew` 같은 하네스 철학을 **출고 검증 한 점에 집중해서** 재해석한 것이다. 범용 orchestration 을 파는 게 아니라, **verify 단계만 떼어내어 productized 한 release harness** 로 보는 편이 훨씬 강하다.

카테고리 핏도 여기서 나온다.
- `Developer Tooling` 으로 설명할 때: "AI-built software 를 위한 release verification layer"
- 설명하지 말아야 할 방식: "여러 모델을 붙일 수 있는 범용 에이전트 플랫폼"

즉, 이 문서의 포인트는 `에이전트가 많다`가 아니라 **"증거 없으면 출고가 안 된다"** 이다.

## 3. 핵심 메커니즘 (3줄)

1. `shiplock-harness` 가 Claude/Codex/Gemini 또는 `oh-my-braincrew` 스타일 파이프라인을 감싸고, 프로젝트 문맥·PR·staging·테스트 결과를 한 번에 수집한다.
2. 수집한 문맥으로 **release contract** 를 자동 생성한 뒤, 요구사항 누락 / 핵심 사용자 플로우 버그 / 기존 합의와의 drift 를 검증한다.
3. 결과를 **standalone HTML evidence report** 로 렌더링해 자동으로 열고, 최종 verdict 를 `PASS / BLOCK / PATCH` 로 반환한다. TUI/GUI 는 같은 verdict 를 다른 채널로 보여주는 부가 인터페이스다.

## 4. 비유로 5분 이해

> 여러 공장에서 만든 부품을 한 컨테이너에 실어 고객 창고로 보내는 물류회사가 있다. 주문서는 이메일, 메신저, 전화, 메모장에 흩어져 있고, 각 공장은 자기 부품만 맞다고 주장한다. 그런데 선적 직전에 **"이 컨테이너 전체가 정말 고객이 주문한 상태인가?"** 를 확인하는 출고 감독이 없다. 그래서 볼트 하나 빠진 상태, 라벨이 틀린 상태, 일부 파손된 상태로 물건이 나가고, 고객은 창고 문 앞에서 컴플레인한다.

shiplock-harness 는 그 **출고 감독 하네스** 다. 여러 공장에서 뭘 만들었는지는 중요하지 않다. 중요한 건 **컨테이너 전체를 출고해도 되는지 마지막 판정을 내리고, 아니면 멈추는 것** 이다.

기술 매핑:
- 여러 공장 = Claude / Codex / Gemini / 멀티에이전트 하네스
- 흩어진 주문서 = Slack / Notion / GitHub / 이슈 / PRD
- 출고 감독 = verification harness
- 출고 확인서 = HTML evidence report

## 5. 라이브 3분 시연 시나리오

**무대 세팅**: cmux 워크스페이스 4개 동시 띄움 — `assistant-run` / `context` / `staging-app` / `report`. 데모용 프로젝트는 "회원가입 후 온보딩 메일 발송" 기능. 실제 고객 요구는 Notion 과 Slack 에 흩어져 있고, 구현은 AI 코딩 어시스턴트가 만든 PR 로 들어와 있다.

| 초 | 화면 | 청중이 보는 것 | 의미 |
|---|---|---|---|
| 0s | 발표자 명령 입력 | `$ shiplock run --assistant codex --pr 42 --staging http://localhost:3000 --report ./out/demo.html --open` | 코딩 어시스턴트를 하네스로 감싼다는 점을 한 줄로 보여줌 |
| 2s | `context` 워크스페이스 | Notion 요구사항 + Slack 합의 + GitHub 이슈를 읽어 release contract 6개 자동 생성 | 원래 정답지가 흩어져 있었음을 먼저 보여줌 |
| 5s | `assistant-run` 워크스페이스 | "all checks green" 처럼 보이는 AI 작업 로그 | AI가 만들었다고 해서 바로 출고 가능한 건 아니라는 대비 |
| 7s | `staging-app` 워크스페이스 | 회원가입은 성공하지만 실제로는 welcome email 미발송, 모바일 CTA 깨짐 | 고객이 나중에 발견할 버그를 눈으로 확인 |
| 10s | 브라우저 자동 open | `demo.html` 리포트가 열리며 상단에 **`VERDICT: BLOCK`** 표시 | 데모의 주인공은 HTML evidence report |
| 13s | 리포트 본문 | failed user flow, missing requirement, patch diff, failing Playwright test 카드 4개 | 단순 경고가 아니라 증거 기반 차단 |
| 18s | GitHub/PR 화면 | merge blocked 코멘트 + check run 실패 | 실제 개발 흐름에서 출고가 멈춤 |
| 24s | 발표자 한 문장 | "AI가 만든 결과물도, 증거 없으면 고객에게 못 나갑니다." | 헤드라인 |

**중요한 데모 원칙**:
- TUI/GUI 를 보여줄 수는 있지만 주연은 아니다.
- 가장 강한 장면은 `open 된 HTML report 의 BLOCK verdict` 다.
- 심사 포인트는 "에이전트 orchestration" 이 아니라 **"버그가 나가기 전에 멈췄다"** 다.

## 6. 9시간 빌드 분해

| 시간 | 마일스톤 |
|---|---|
| 08:00–08:30 | 환경 셋업 (`uv`, `playwright`, `rich`, HTML report 템플릿, cmux 동작 확인) |
| 08:30–09:30 | `shiplock` CLI 스캐폴딩 + assistant adapter 인터페이스 정의 (`claude` / `codex` / `gemini`) |
| 09:30–10:30 | 문맥 수집기 구현 (Slack/Notion/GitHub 대신 데모용 로컬 fixtures ingest) |
| 10:30–11:30 | release contract 생성기 + risk scoring 로직 |
| 11:30–12:30 | staging 플로우 검사 (Playwright 핵심 사용자 여정 2~3개) |
| 12:30–13:00 | 점심 + 버그 시드 심기 |
| 13:00–14:00 | HTML evidence report 생성기 구현 (`verdict`, `evidence`, `patch`, `tests`) |
| 14:00–15:00 | `--open` 플로우 연결 + PR block 출력 |
| 15:00–16:00 | 단일 어시스턴트 실제 연동 + 나머지 어댑터는 stub 또는 config 예시로 남김 |
| 16:00–17:00 | cmux 4워크스페이스 데모 시퀀스 연결 + 카피 polish |
| 17:00–18:00 | 최종 리허설 + fallback 루트 점검 |

**라이브러리**:
- 실행: `playwright`, `sqlite3`, `rich`
- LLM 호출: `google-genai` 또는 `vertexai`
- 리포트: self-contained HTML 생성 스크립트 (`jinja2` 또는 단일 Python renderer)

**해커톤 범위 제한**:
- 실제로는 `assistant-agnostic` 을 지향하지만, 데모 구현은 **단일 실제 어댑터 + 나머지 인터페이스 명세** 만 보여줘도 충분하다.
- 진짜 제품은 TUI/GUI 확장 가능하지만, 해커톤에서는 **HTML report 하나만 완성도 높게** 만드는 것이 맞다.

## 7. 리스크 + 대응 + 채점

**3축 채점**: ① 한 컷 시연 10/10 / ② 헤드라인 9/10 / ③ 카테고리 적합성 10/10 — **합 29/30**

**리스크**:
- (높음) `범용 하네스`처럼 들리면 문제 정의가 흐려짐
- (중) 여러 assistant adapter 를 다 구현하려다 범위가 터질 수 있음
- (중) HTML report 가 예쁘기만 하고 "왜 block 인지"가 약하면 데모가 쇼처럼 보일 수 있음
- (낮음) 외부 SaaS 실제 연동까지 가면 셋업 시간이 길어짐

**대응**:
- 발표 문장과 문서에서 항상 `verification harness`, `release contract`, `evidence report`, `block` 네 단어를 중심축으로 유지
- 멀티 어시스턴트는 구조만 열어두고, 해커톤 구현은 `Codex 또는 Claude 1개`만 실제 연결
- 리포트 최상단을 무조건 `VERDICT / WHY BLOCKED / FIX NOW` 3단 구조로 고정
- 외부 연동은 fixture 로 대체하고, 핵심은 `문맥 -> verdict -> open report -> merge blocked` 시퀀스를 살리는 데 집중

**Fallback**:
- 브라우저 자동 open 이 불안정하면 HTML 파일 경로를 출력하고 수동 `open` 으로 전환
- PR 연동이 늦어지면 `shiplock run` 단일 CLI 와 HTML report 만으로 시연
- 멀티 어시스턴트 wrapper 가 늦어지면 `--assistant codex` 한 가지 모드만 남기고, 나머지는 config 예시로 설명

**가장 큰 강점**:
- "모델이 똑똑하다"를 팔지 않는다. **모델이 만든 결과물을 누가 출고 판정하느냐** 를 판다.
- "에이전트가 많다"를 팔지 않는다. **증거 없이 출고되지 못하게 막는 하네스** 를 판다.
- 새 세션에서 이 문서를 읽는 모델은 반드시 이 후보를 `범용 플랫폼`이 아니라 **Developer Tooling용 release verification harness** 로 이해해야 한다.
