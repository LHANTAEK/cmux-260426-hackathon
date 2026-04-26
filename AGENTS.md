# CLAUDE.md

이 파일은 Claude Code가 본 디렉토리에서 작업할 때 따를 컨텍스트와 규칙을 정의한다.

## Project Overview

**CMUX × AIM Intelligence 해커톤 (2026-04-26 일요일, 8am–6pm, Monacospace)** 단일 최종 제출안 선정을 위한 브레인스토밍 및 제안서 워크스페이스.

- 형식: 터미널 only, IDE 금지, AI coding agent로 빌드
- 인프라: Gemini + GCP 크레딧 $40,000 제공
- 심사: AIM Intelligence 엔지니어 (Samsung Ventures $7M Series A, Stinger·Starfort 운영사)
- 인원: Solo or 팀 (최대 4명)

본 디렉토리는 **카테고리별 후보를 같은 톤·포맷으로 병렬 propose** 한 뒤, 시각적 비교로 최종 1개를 선택하기 위한 의사결정 워크스페이스다.

### 해커톤 3개 카테고리 (나중에 제출은 1개만 선택)

1. **AI Safety & Security** — Find, expose, and defend against AI vulnerabilities. Red teaming CLIs, guardrail pipelines, jailbreak detection, prompt injection scanners, LLM fuzzing.
2. **Developer Tooling** — Build tools that make AI-assisted development faster, safer, or more observable.
3. **Business & Applications** — Ship user-facing AI products that solve real business problems.

## Decision Principle (사용자 명시)

> "복잡하지 않고, 핵심 문제를 발견해서 심플하게 해결한다."
> "문제 발견이 굉장히 중요해."

- 단순한 프레임: **문제 발견 → 해결**
- 멋있는 기술 스택보다 "한 컷 시연 + 한 줄 헤드라인 + 심사관 직결"의 3축으로 채점한다.

## Current Working Direction (2026-04-26)

현재 가장 강한 `Developer Tooling` 축은 **`release-harness-on-cmux`** 다.

- 문제 정의: 작은 AI 팀은 QA/QC 인력이 없고, 고객별 성공 기준이 Slack·Notion·GitHub 에 흩어져 있어 `잘 돌아가는 데모`를 `맞는 출고물`로 착각한 채 출고한다.
- 제품 포지셔닝: **AI coding assistant 기능이 아니라, 어떤 assistant / agent / workflow 바깥에도 붙일 수 있는 customer-aware release harness** 로 본다.
- 핵심 입력: `Slack thread`, `Notion page/export`, `GitHub PR diff`, 필요하면 `MCP config`, `prompt log`, `staging URL`
- 핵심 출력: `Go / No-Go 판정`, 실패 근거 3개, screenshot/diff/evidence, 자동 생성된 `report.json`, `report.html`
- 기본 실행 표면: `CLI first`
- 기본 시연 표면: `cmux alert + HTML report`
- 보조 시각화: 얇은 `TUI`
- 금지할 과욕: 해커톤 MVP에서 full GUI platform 을 만들려고 하지 않는다. **본체는 `CLI + HTML report + open`** 이고, TUI/GUI 는 후순위다.

### Harness Reference

아키텍처 감각은 `/Users/limhantaek/oh-my-braincrew/oh-my-braincrew` 를 참조한다.  
다만 **그 저장소 전체를 베끼는 것이 목적이 아니다.**

빌릴 패턴:
- Go 또는 단일 CLI entrypoint 중심의 하네스 구조
- adapter / runner / report renderer 분리
- hook 또는 wrapper 로 외부 workflow 에 붙는 방식
- state/evidence 를 파일로 남기고, opt-in 시각화 레이어를 따로 두는 방식

이번 해커톤에서 필요한 범위:
- `spec compiler`: 고객 성공 기준을 `customer_contract.json` 또는 `spec.yaml` 로 컴파일
- `runner layer`: scenario check, hard-rule check, Playwright smoke, 좁은 범위의 LLM judge
- `evidence store`: 실패 이유, 인용 문장, screenshot, diff
- `report renderer`: `report.html`
- `launcher`: `client-release approve ... --open-report`

### Killer Demo Constraint

가장 중요한 데모는 하나다.

- 고객사 `acme-bank`
- 요구사항: `CSV export 필수`, `white-label only`, `enterprise tone`
- 현재 데모의 실패: `beta_banner=true`, export route 없음, 보고서 tone drift
- 발표 장면: CLI 실행 -> cmux 연쇄 점멸 -> `report.html` 자동 오픈 -> `BLOCKED`

새 세션에서는 `release-harness-on-cmux` 를 논의할 때 반드시 위 문제정의와 MVP 범위를 유지한다.

## Folder Layout

```
cmux-260426-hackathon/
├── CLAUDE.md                              # 본 파일
├── dev-list.md                            # 후보 인덱스 (사용자 작성)
├── README2.md                             # specfirewall 초안
├── README3.md                             # shiplock 초안
├── README4.md                             # release-harness-on-cmux 초안 (현재 Developer Tooling 1순위)
├── README5.md                             # demo-preflight-harness 초안
└── docs/
    └── 01-mcp-rugcheck-on-cmux/           
        └── proposal.md
```

## Proposal Tone & Template

각 후보의 `proposal.md` 는 동료 pangpang@brain-crew.com 에게 보낸 메일 v2 의 톤앤매너를 그대로 따른다. **모든 후보에 동일한 7섹션** 을 둔다.

```markdown
# [코드네임] — [한 줄 헤드라인]

## 1. 풀려는 문제 정의
한 문장으로 핵심 위험을 정의 + 근거 (CVE / 사건 / 보고서 출처).

## 2. 왜 지금 우리가 풀어야 하는가
AIM Intelligence 의 어떤 공개 repo·제품과 1:1 매칭되는지.
"심사관이 자기 도구 위에서" 시연되는 효과의 근거.

## 3. 핵심 메커니즘 (3줄 이내)
입력 → 처리 → 출력. 라이브러리 1~2개로 끝나야 한다.

## 4. 비유로 5분 이해
비전공자 / 비도메인자 / 동료에게 5분 안에 납득시킬 수 있는 비유.

## 5. 라이브 3분 시연 시나리오
표 형식 — `초 | 화면 | 청중이 보는 것 | 의미`.
cmux 어떤 워크스페이스에서 어떻게 보일지 명시.

## 6. 9시간 빌드 분해
시간별 마일스톤 (8am–6pm). 점심·시연 리허설 포함.
사용 라이브러리·SDK·템플릿 확정.

## 7. 리스크 + 대응 + 채점
3축 (① 한 컷 시연 / ② 헤드라인 / ③ 심사관 직결) 점수 + 빌드 리스크 + fallback.
```

## Reference Context (본 의사결정의 근거 자료)

### AIM Intelligence 공개 repo (= 심사관 정답지)
- AIM-MCP, awesome-mcp-security → MCP 보안
- WhisperInject (audio jailbreak), AIM-Robotics (Unitree G1), video2robot → 멀티모달
- Automated-Multi-Turn-Jailbreaks → 멀티턴 우회
- AIM-Forge, auto-aim → 자동화 red-teaming
- 제품: Stinger (자동 red teaming) + Starfort (실시간 guardrail), "spear and shield"

### 핵심 사건·CVE (2026.02 ~ 2026.04)
- 2026.04.16 *The Register*: Anthropic MCP SDK "by design" RCE, 200,000 인스턴스 영향
- 2026.02 *OX Security*: LobeHub·Cursor Directory PoC 페이로드 11개 업로드, 9개가 보안 리뷰 0회로 통과
- 2026.03 *Agent Commander*: coding agent를 C2 채널로 변환 시연
- CVE-2025-59536, CVE-2026-21852 (Claude Code 프로젝트 파일 trust dialog 전 실행)
- CVE-2026-21516 (CVSS 9.6)
- CVE-2026-22252, CVE-2026-25724
- AudioJailbreak (arXiv 2505.14103), Multi-AudioJail, RIR 시뮬레이션

### cmux 통합 메커니즘
- OSC 9 escape sequence 한 줄 (`printf '\033]9;message\033\\'`) 으로 cmux 사이드바 알림 점멸
- 워크스페이스 다중 동시 띄우기 → 4개 사이드바 연쇄 점멸이 가장 강한 시각 신호

## Working Rules

- 모든 propose.md 는 한국어로 작성
- 시각적 시연 묘사는 표 또는 ASCII 다이어그램으로 — 5초 안에 청중이 알 수 있어야 함
- 후보 간 헤드라인 중복 금지 (메일 v2 의 "한 줄 헤드라인" 룰)
- 9시간 빌드 분해는 반드시 8am 시작, 6pm 시연으로 끊는다
- 새 후보 추가 시 `dev-list.md` 와 본 CLAUDE.md 의 Folder Layout 동시 갱신
- 현재 1순위 후보를 다룰 때는 `generic harness` 가 아니라 `release harness` 로 좁혀 설명한다
- `cmux` 는 본체가 아니라 시연 채널, `HTML report` 는 증빙 산출물이라는 구분을 유지한다
