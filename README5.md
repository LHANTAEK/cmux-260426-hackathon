# 05. demo-preflight-harness — "고객이 실제로 쓸 장면이 빠진 데모는 출고 전에 걸린다"

## 1. 풀려는 문제 정의

**"작은 AI 팀은 고객별 성공 기준이 이메일·노션·깃허브에 암묵지로 흩어진 채 데모·웹 기능·LLM 산출물을 출고하고, QA/QC 인력이 없어 고객이 먼저 '우리가 원한 건 이게 아닌데요?'를 말한다."**

근거:
- 2026.04.15 *OX Security*: MCP 공급망 딥다이브는 이제 **숨은 설정과 컨텍스트 자체가 실행 레이어**가 됐음을 보여줬다. 출고 리스크는 더 이상 UI 버그 하나로 끝나지 않는다
- 2026.03.16 *Embrace The Red* `Agent Commander`: 자연어 수준의 은근한 지시만으로도 agent tasking 이 바뀔 수 있다. 즉, **맥락 미스매치 자체가 운영 리스크**다
- AIM Intelligence 공개 리서치 `Tool-Mediated Belief Injection`: tool output 이 만든 잘못된 전제가 이후 판단 전체에 누적될 수 있다. **겉으로는 멀쩡한 데모도 고객 맥락에서는 이미 틀릴 수 있다**
- 2025.08.02 EU GPAI 의무 적용, 2026.01.22 한국 AI 기본법 발효. 시장은 이미 **출고 전 리스크 관리와 human oversight** 를 요구하는 쪽으로 움직이고 있다

## 2. 왜 지금 우리가 풀어야 하는가

심사 = AIM Intelligence. 그들의 언어는 이미 정해져 있다.  
`Stinger` 는 custom vulnerability testing, `Starfort` 는 real-time adaptive guardrail 이다.

demo-preflight 는 그 사이 빈칸을 메운다. 우리는 generic red teaming 을 더 하는 것이 아니라, **고객별 성공 기준을 출고 전 fail/pass 게이트로 컴파일** 한다.

즉:
- `Stinger` 쪽으로 보면: 고객별 failure mode 를 출고 전에 먼저 깨보는 장치
- `Starfort` 쪽으로 보면: 운영 전에 이미 잘못된 행동을 막는 pre-production guardrail

여기서 중요한 포지셔닝은 **"AI 코딩 어시스턴트용 기능"이 아니라 "assistant 바깥에 붙는 harness"** 라는 점이다.  
즉, 특정 IDE·모델·에이전트에 종속되지 않는다. `이메일 / 노션 / GitHub / MCP / staging app / prompt log` 중 무엇이든 adapter 로 붙이고, 결과는 공통 runner 와 report renderer 가 처리한다.

또한 시연 무대가 cmux 자체다. 보통의 QA 도구는 CI 로그 한 줄로 끝나지만, 우리는 **어느 고객 문맥에서 무엇이 어긋났는지**를 cmux 워크스페이스 4개와 자동으로 열리는 HTML report 에서 동시에 보여줄 수 있다.  
즉, `Developer Tooling` 이면서도 **심사장에서 한 컷으로 박히는 도구**가 된다.

## 3. 핵심 메커니즘 (3줄)

1. `이메일 thread + Notion page + GitHub PR` 을 읽어 고객별 성공 기준·금지사항·승인 흐름을 **client_spec.json** 으로 컴파일
2. harness 가 `staging URL + report.md + config.json` 에 runner 를 붙여 핵심 사용자 플로우 누락, 약속 충돌, tone drift 를 **demo risk** 로 판정하고 evidence 를 저장
3. 위험 점수가 기준을 넘으면 `demo-preflight run --open` 이 출고를 차단하고, cmux OSC 9 알림 + **자동 생성된 `report.html` 오픈**으로 실패 근거를 즉시 보여준다

## 4. 비유로 5분 이해

> 작은 행사 대행사가 오늘 4개 고객사 데모를 연속으로 보여줘야 한다. 고객 요구는 여기저기 흩어져 있다. 메일에는 "반드시 관리자 승인 후 발송", 노션에는 "beta 문구 금지", 깃에는 새 설정값이 올라와 있다. 팀은 적고 바쁘다 보니 데모는 그럴듯하게 돌아가지만, **고객이 실제로 쓰는 장면 하나가 빠진 채 발표장으로 들어간다.** 고객은 데모가 끝나고 나서야 말한다. "우리가 쓰는 흐름은 그게 아닌데요?"

demo-preflight 는 발표장 문 앞에 서 있는 **리허설 감독**이다. 화면이 예쁜지보다, **이 고객이 실제로 쓰는 장면이 재현됐는지**를 시작 직전에 막아 세운다.
감독은 특정 배우만 상대하지 않는다. 연극팀이 바뀌어도, 대본 형식이 바뀌어도, **리허설 체크리스트와 사고 리포트 포맷**만 같으면 계속 쓸 수 있다.

기술 매핑:
- 고객과의 대화 = 이메일·노션·깃허브에 흩어진 컨텍스트
- "우리가 실제로 쓰는 장면" = 고객별 성공 기준
- 리허설 감독 = demo-preflight harness
- 발표장 빨간 경광등 = cmux OSC 9 연쇄 점멸

## 5. 라이브 3분 시연 시나리오

**무대 세팅**: cmux 워크스페이스 4개 동시 띄움 — `email-thread` / `notion-spec` / `github-pr` / `release-lane`. 브라우저에는 비어 있는 `report.html` 탭 하나만 미리 열어 둔다.  
데모용 고객사 `acme-bank` 는 3개 요구사항을 가진다: `상담원 초안 -> 관리자 승인 -> 고객 발송`, `white-label only`, `CSV export 필수`.

| 초 | 화면 | 청중이 보는 것 | 의미 |
|---|---|---|---|
| 0s | 발표자 명령 입력 | `$ ./demo-preflight run acme-bank --pr 142 --notion ACM-12 --email fixtures/acme.eml --open` | 하네스 실행 + 리포트 자동 오픈 |
| 1s | `release-lane` 로그 | `collecting customer context...` | 단순 QA 가 아니라 고객 문맥 수집부터 시작 |
| 2s | `email-thread` 하이라이트 | `상담원이 초안 작성 후 관리자 승인 거쳐야 합니다` 노란 강조 | 고객의 실제 사용 시나리오 원문 |
| 3s | `notion-spec` 하이라이트 | `No beta badge / white-label only` 빨간 밑줄 | 문서상 금지사항 |
| 4s | `github-pr` diff | `auto_send=true`, `beta_banner=true`, export route 없음 | 코드와 고객 문맥이 충돌 |
| 5s | `staging` 요약 캡처 | 승인 단계 없이 바로 발송되는 화면 | "잘 돌아가는 데모" 와 "맞는 데모" 는 다름 |
| 6s | **cmux 사이드바 4개 빨강 연쇄 점멸** | 시각적 폭발 | "출고 중지"를 한 컷으로 전달 |
| 8s | 브라우저 `report.html` | 상단에 `BLOCKED`, 아래에 실패 시나리오 3개와 근거 문장 3개 | HTML report 가 심사장용 산출물이라는 점을 강조 |
| 10s | `release-lane` 결과 | `BLOCKED: missing approval step / beta badge exposed / CSV export missing` | 고객 시나리오 미스매치를 근거와 함께 설명 |
| 14s | 발표자 한 문장 | "우리는 assistant 를 더 만드는 게 아니라, 어떤 assistant 든 바깥에서 검증하고 출고를 막는 하네스를 만듭니다" | 헤드라인 |

## 6. 9시간 빌드 분해

| 시간 | 마일스톤 |
|---|---|
| 08:00–08:30 | 환경 셋업 (`uv`, `pydantic`, `rich`, `jinja2`, `playwright`, `google-genai`, cmux OSC 9 확인) |
| 08:30–09:30 | `이메일 .eml / Notion export / GitHub PR diff` fixture loader 3종 |
| 09:30–11:00 | 고객사 성공 기준·금지사항 추출 → `client_spec.json` 생성 |
| 11:00–12:00 | demo bundle adapter (`report.md`, `config.json`, `staging URL`) + evidence 저장 포맷 정의 |
| 12:00–12:30 | 하드룰 검사기 + 좁은 범위의 LLM judge (tone / omission 전용) |
| 12:30–13:00 | 점심 + 데모용 fixture 시드 (`acme-bank`) |
| 13:00–14:30 | `demo-preflight run --open` CLI + block/allow flow 완성 |
| 14:30–15:30 | Playwright 핵심 사용자 플로우 smoke test + screenshot 캡처 |
| 15:30–16:00 | `report.html` 렌더러 + `open` launcher 연결 |
| 16:00–16:30 | cmux OSC 9 연동 + 빨간 diff 시각화 |
| 16:30–17:30 | E2E 시연 리허설 ×3 |
| 17:30–18:00 | fallback 준비 + 발표 슬라이드 1장 |

**라이브러리**: `pydantic`, `rich`, `jinja2`, `playwright`, `google-genai`, `sqlite3`

## 7. 리스크 + 대응 + 채점

**3축 채점**: ① 한 컷 시연 9/10 / ② 헤드라인 10/10 / ③ 심사관 직결 9/10 — **합 28/30**

**리스크**:
- (중) 이메일·노션·깃허브 live API 연동은 토큰·권한·속도 이슈가 있다 → 해커톤 빌드는 **exported fixture ingestion** 으로 고정
- (중) LLM judge 가 흔들릴 수 있다 → 판정의 80% 는 하드룰과 시나리오 누락 검사로 두고, LLM 은 tone / omission 만 담당
- (낮음) staging smoke test 가 flaky 할 수 있다 → Playwright 실패 시에도 spec mismatch 와 config/report 검사만으로 block 사유는 유지
- (낮음) TUI/GUI 욕심이 커지면 본체가 흐려질 수 있다 → MVP 는 **CLI + HTML report + open**, TUI/GUI 는 후순위

**Fallback**: live 연동이 늦어지면 `fixtures/acme.eml`, `fixtures/notion.md`, `fixtures/pr.diff` 로 다운그레이드. 브라우저 자동 오픈이 불안하면 `report.html` 경로만 출력하고 수동으로 연다. 본체는 그대로 `고객 시나리오 미스매치 차단` 이고, cmux 점멸과 실패 근거 시연은 유지된다.

**가장 큰 강점**: 이 제안은 지식관리 툴이 아니다. 보고서 툴도 아니다. **고객이 화내기 전에 마지막 3분 안에 출고를 막는 pre-complaint defense harness** 다. 작은 AI 팀, 많은 컨텍스트, 부족한 QA/QC 라는 현실 문제를 가장 직접적으로 찌르면서도, 특정 assistant 에 종속되지 않아 `Developer Tooling` 카테고리 설명력이 높다.
