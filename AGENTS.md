# CLAUDE.md

이 파일은 Claude Code가 본 디렉토리에서 작업할 때 따를 컨텍스트와 규칙을 정의한다.

## Project Overview

**CMUX × AIM Intelligence 해커톤 (2026-04-26 일요일, 8am–6pm, Monacospace)** AI Safety & Security 트랙 출품 준비.

- 형식: 터미널 only, IDE 금지, AI coding agent로 빌드
- 인프라: Gemini + GCP 크레딧 $40,000 제공
- 심사: AIM Intelligence 엔지니어 (Samsung Ventures $7M Series A, Stinger·Starfort 운영사)
- 인원: Solo or 팀 (최대 4명)

본 디렉토리는 **5개 후보를 같은 톤·포맷으로 병렬 propose** 한 뒤, 시각적 비교로 1개를 선택하여 클로드 코드를 활용한 4시간 빌드에 들어가기 위한 의사결정 워크스페이스다.

### 해커톤 3개 카테고리 (나중에 제출은 1개만 선택)

1. **AI Safety & Security** — Find, expose, and defend against AI vulnerabilities. Red teaming CLIs, guardrail pipelines, jailbreak detection, prompt injection scanners, LLM fuzzing.
2. **Developer Tooling** — Build tools that make AI-assisted development faster, safer, or more observable.
3. **Business & Applications** — Ship user-facing AI products that solve real business problems.

## Decision Principle (사용자 명시)

> "복잡하지 않고, 핵심 문제를 발견해서 심플하게 해결한다."
> "문제 발견이 굉장히 중요해."

- 단순한 프레임: **문제 발견 → 해결**
- 멋있는 기술 스택보다 "한 컷 시연 + 한 줄 헤드라인 + 심사관 직결"의 3축으로 채점한다.

## Folder Layout

```
cmux-260426-hackathon/
├── CLAUDE.md                              # 본 파일
├── dev-list.md                            # 후보 인덱스 (사용자 작성)
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
