# 02. specfirewall — "고객이 아직 말로 다 못 한 성공 기준을, 배포 전 fail/pass 게이트로 바꾼다"

## 1. 풀려는 문제 정의

**"고객별 성공 기준이 메일·노션·메신저·깃허브에 암묵지로 흩어진 채 AI 데모·에이전트·웹서비스가 출고되면, 팀은 '잘 돌아가는 데모'를 납품하지만 고객은 '우리가 원한 건 이게 아닌데요?'를 말하게 된다."**

근거:

- 2026.04.15 *OX Security* "The Mother of All AI Supply Chains": 숨은 MCP 설정과 컨텍스트가 실행 권한으로 바로 이어지며, **"명시되지 않은 전제"** 가 곧 사고 surface 가 됨
- 2026.03.16 *Embrace The Red* `Agent Commander`: 에이전트는 자연어 수준의 은근한 지시만으로도 공격자 목적에 맞게 tasking 될 수 있음 → **맥락 오염** 자체가 리스크
- 2026년 AIM Intelligence 공개 리서치 `Tool-Mediated Belief Injection`: tool output 이 만든 잘못된 전제가 대화 전체에 누적될 수 있음
- OpenAI `Evaluation best practices` (2026): production data + historical data + explicit metric 없이 퍼지한 목표로는 안정적인 출고 품질을 만들기 어려움

즉, 이 문제는 "테스트가 부족하다"가 아니라 **고객 성공 기준이 테스트 가능한 형태로 번역되지 않은 채 출고된다** 는 문제다.

## 2. 왜 지금 우리가 풀어야 하는가

심사 = AIM Intelligence. 그들의 현재 제품 언어는 매우 선명하다.

- `Stinger`: business logic 기반 custom vulnerability testing
- `Starfort`: real-time adaptive guardrails
- AIM 홈페이지: **development to production** 전 주기 통제

specfirewall 은 이 둘의 사이 빈칸을 메운다.  
우리는 generic red teaming 을 더 하는 게 아니라, **고객별 성공 기준을 custom release gate 로 컴파일** 한다.

즉:

- Stinger 쪽으로 보면: 고객별 failure mode 를 자동 발견하는 사전 red teaming
- Starfort 쪽으로 보면: 배포 전부터 고객 요구 위반을 막는 pre-production guardrail

또한 시연 무대가 cmux 자체다. **심사관이 자기 도구 감각으로 바로 이해할 수 있는 "출고 전 차단"** 장면이 나온다.  
generic AI QA 가 아니라, AIM 의 "spear and shield" 메타포를 **고객 성공 기준** 레이어까지 확장하는 셈이다.

## 3. 핵심 메커니즘 (3줄)

1. 고객 메일·노션·깃허브 이슈·실제 로그를 읽어 `성공 기준 후보` 를 추출하고, 사람이 30초 안에 승인/수정한다
2. 이를 `spec.yaml` 로 구조화한 뒤 agent eval / web flow test / load check 로 자동 컴파일한다
3. 배포 직전 `pass/fail + 근거 evidence report` 를 출력하고, 실패 시 cmux OSC 9 알림으로 워크스페이스를 즉시 점멸시킨다

## 4. 비유로 5분 이해

> 맞춤 정장을 만든다고 하자.  
> 손님 A 는 "핏" 이 중요하고, 손님 B 는 "하루 안에 받아야" 하고, 손님 C 는 "안쪽 포켓이 꼭 있어야" 한다.  
> 그런데 재단사는 "어쨌든 입을 수는 있잖아" 라는 기준으로 출고한다. 옷은 멀쩡하지만, 고객은 바로 컴플레인한다.

specfirewall 은 재단사를 더 열심히 재봉하게 만드는 도구가 아니다.  
그보다 먼저, **"이 손님에게 정장 완성의 의미가 뭐냐"** 를 체크리스트로 박아 넣는 출고 검사표다.

기술 매핑:

- 손님과 나눈 대화 = 메일 / 노션 / 메신저 / 깃허브
- "핏 / 속도 / 포켓" = 고객별 성공 기준
- 출고 검사표 = `spec.yaml`
- 정장 출고 직전 최종 확인 = agent/web/load QA gate
- 점멸 경고 = cmux OSC 9

## 5. 라이브 3분 시연 시나리오

**무대 세팅**: cmux 워크스페이스 4개 동시 띄움 — `mail-context` / `spec-compiler` / `agent-eval` / `web-load-check`. 고객 A 의 메일 스레드, 노션 요구사항 요약, 웹서비스 staging URL, 에이전트 시나리오 fixture 준비.


| 초   | 화면                  | 청중이 보는 것                                                                                    | 의미                        |
| --- | ------------------- | ------------------------------------------------------------------------------------------- | ------------------------- |
| 0s  | 발표자 메일 열기           | 고객 메일 하이라이트: "정확도보다 감사로그가 중요", "동시 50명에서도 안 터져야 함"                                          | 성공 기준은 코드 밖에 흩어져 있음       |
| 5s  | 명령 입력               | `$ specfirewall compile ./fixtures/customer-a`                                              | 맥락을 테스트 가능한 기준으로 컴파일 시작   |
| 8s  | `spec.yaml` 출력      | `audit_log=required`, `latency_p95<2.5s`, `concurrency>=50`, `agent_must_cite_sources=true` | 암묵 요구의 구조화                |
| 12s | agent eval 실행       | 핵심 답변은 맞지만 출처 미표시 → `FAIL`                                                                  | "잘 답했다" 와 "고객 기준 충족" 은 다름 |
| 16s | web/load check 실행   | UI 플로우 통과, 하지만 50명 동접에서 에러율 상승 → `FAIL`                                                     | 기능 정상과 출고 가능은 다름          |
| 20s | **cmux 사이드바 빨강 점멸** | `agent-eval`, `web-load-check` 워크스페이스 2개 동시 점멸                                              | 한 컷 시연의 정점                |
| 24s | evidence report     | `배포 불가 사유 2개 + 근거 로그 2개 + 고객 문장 원문 2개`                                                      | 납득 가능한 설명                 |
| 35s | 발표자 한 문장            | "우리는 버그를 찾는 게 아니라, 고객이 기대한 성공 기준을 못 맞춘 출고를 막습니다"                                            | 헤드라인                      |


## 6. 9시간 빌드 분해


| 시간          | 마일스톤                                                                  |
| ----------- | --------------------------------------------------------------------- |
| 08:00–08:30 | 환경 셋업 (`uv`, `pydantic`, `rich`, `playwright`, `locust`) + fixture 정리 |
| 08:30–10:00 | 메일/노션/깃허브 text fixture ingestion + 성공 기준 추출 프롬프트                      |
| 10:00–11:30 | `spec.yaml` 스키마 정의 + human approve/edit CLI                           |
| 11:30–12:30 | agent eval runner (정확도, 출처 표시, 포맷, 금지 응답)                             |
| 12:30–13:00 | 점심                                                                    |
| 13:00–14:00 | web flow runner (`playwright`) + 핵심 플로우 pass/fail                     |
| 14:00–15:00 | load check runner (`locust` 또는 간단한 async burst test)                  |
| 15:00–16:00 | evidence report 생성 + `배포 가능/불가` 최종 판정                                 |
| 16:00–17:00 | cmux OSC 9 통합 + 시연 스크립트 polish                                        |
| 17:00–18:00 | 리허설 ×3 + 발표 문구 고정                                                     |


**라이브러리**: `pydantic`, `rich`, `playwright`, `locust`, `sqlite3`  
**핵심 구현 원칙**: 실제 외부 연동 대신 exported text fixture 로 고정해, extraction 과 gate logic 에 집중

## 7. 리스크 + 대응 + 채점

**3축 채점**: ① 한 컷 시연 9/10 / ② 헤드라인 10/10 / ③ 심사관 직결 10/10 — **합 29/30**

**리스크**:

- (중) 암묵 요구 추출이 과하게 fuzzy 할 수 있음  
→ 대응: 완전 자동이 아니라 `human approve/edit` 30초 단계를 넣어 권위 확보
- (중) agent / web / load 3개 runner 를 다 실물로 붙이면 시간이 빠듯함  
→ 대응: **공통 spec 엔진 하나** 를 만들고, runner 는 2개 실물 + 1개 단순 metric check 로 축소 가능
- (낮음) 메일·노션·깃허브 실연동이 데모 안정성을 해칠 수 있음  
→ 대응: exported fixture 로 고정하고 "실서비스 연결 가능" 메시지만 남김

**Fallback**:

- 시간이 부족하면 `hero input = 고객 메일`, `hero outputs = agent eval + readiness report` 만으로 축소
- web/load runner 는 mock metric 로 대체해도 본질은 유지됨

**가장 큰 강점**:

- 흔한 "AI QA 툴" 이 아니라 **고객 성공 기준 컴파일러** 라는 비대칭 framing
- AIM 의 `custom vulnerability testing` 과 `adaptive guardrail` 사이를 메우는 아이디어
- 실제 작은 AI 조직 pain 과 심사관의 AI safety 언어가 한 점에서 만남

