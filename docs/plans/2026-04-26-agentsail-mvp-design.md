# Agent Sail MVP — 설계 문서 (2026-04-26)

CMUX × AIM Intelligence 해커톤 1순위 후보 Agent Sail의 9시간 빌드 가능한 MVP 정의.

## 1. 한 줄 정의

> 같은 에이전트, 다른 고객, 다른 출시 게이트. Agent Sail은 고객별 출시 기준을 실행 가능한 증거 기반 게이트로 바꾼다.

세부 사양은 `docs/02-agent-sail/proposal.ko.md`, `AGENTS.md` 가 진실의 원천. 본 문서는 9시간 빌드 범위의 골격만 정의한다.

## 2. 형태 — oh-my-braincrew harness 골격을 그대로 따른다

`/Users/limhantaek/oh-my-braincrew/oh-my-braincrew/` 의 4개 구성을 1:1 매핑.

| oh-my-braincrew | Agent Sail |
|---|---|
| `cmd/omb/main.go` Go binary | `cmd/agentsail/main.go` Go binary |
| `.claude-plugin/plugin.json` + `.claude/{agents,skills,commands,hooks,rules}/` | 동일 구조, namespace 만 `agentsail` |
| `.omb/{plans,executions,verifications,…}/` 파일 evidence store | `.agentsail/{cache,contracts,runs,reports}/` |
| Hotl state machine + hook 자동 step 전환 (6단계) | **본 MVP 범위 외**. 단순 sequential 4단계. |

oh-my-braincrew 의 6단계 구현 파이프라인 → Agent Sail의 4단계 출시 게이트로 좁힌다.

```
oh-my-braincrew:  plan → review → execute → verify → document → pr     (구현 파이프라인)
agent-sail   :  collect → compile → check → verdict → report           (출시 게이트)
                ^^^^^^^   --------------- agentsail ci ---------------
                Claude Code skill        Go binary 단일 호출 sequential
```

## 3. 가정 (state assumptions — Karpathy 1)

- 검사 대상 target 은 사용자가 시연 시점에 별도 localhost 서버로 띄운다. 본 레포에는 `examples/http_target/main.go` Go HTTP mock chat endpoint 만 둔다.
- 외부 SaaS auth 는 Claude Code MCP / `gh` CLI 가 처리. **`.env` 사용하지 않는다.**
- LLM judge (omission/tone) 가 필요한 좁은 검사는 **Claude Code subagent** 에 위임. 별도 Gemini/Anthropic SDK 호출 안 한다.
- `langgraph` adapter 는 본 MVP 범위 외. proposal §11의 3개 adapter 중 `mock` + `http` 만 구현.
- Hotl state machine + hook 자동 step 전환 + Playwright smoke + 실시간 load probe + rich.Live TUI 는 다음 세션.

## 4. 명령 인터페이스

### Claude Code skill (collect 단계)

```
/agentsail collect <customer>
```

- 7개 collector agent (`slack/notion/email/github/staging/test/assistant-log`) 병렬 fanout
- MCP (`mcp__slack__*`, `mcp__claude_ai_Notion__*`, `mcp__claude_ai_Gmail__*`) + `gh` CLI 로 raw context 수집
- `.agentsail/cache/<customer>/{slack,gmail,notion,github}.json` 으로 떨굼

### Go binary (compile → check → verdict → report)

```
agentsail ci --customer <X> --target http://localhost:PORT/path [--report] [--open]
agentsail report .agentsail/runs/<run-id>.json [--open]
agentsail version
agentsail doctor
```

- cache 가 비어 있으면 `Run /agentsail collect <customer> first` 로 즉시 종료
- 단계 사이마다 cmux OSC9 alert 한 줄 출력 (`printf '\033]9;...\033\\'`)

## 5. 파일 트리

```
cmux-260426-hackathon/
├── .claude-plugin/plugin.json
├── .claude/
│   ├── agents/agentsail/
│   │   ├── slack-collector.md         # mcp__slack__*
│   │   ├── notion-collector.md        # mcp__claude_ai_Notion__*
│   │   ├── email-collector.md         # mcp__claude_ai_Gmail__*
│   │   ├── github-collector.md        # gh CLI
│   │   ├── contract-compiler.md       # raw context → customer_contract.json
│   │   ├── criteria-checker.md        # deterministic check
│   │   ├── chaos-prober.md            # 429 / timeout / empty retrieval
│   │   ├── verdict-engine.md          # SHIP/HOLD/BLOCK
│   │   └── report-renderer.md         # HTML evidence
│   ├── skills/
│   │   ├── agentsail/SKILL.md          # 디스패처
│   │   ├── agentsail-collect/SKILL.md
│   │   ├── agentsail-compile/SKILL.md
│   │   ├── agentsail-check/SKILL.md
│   │   ├── agentsail-verdict/SKILL.md
│   │   ├── agentsail-report/SKILL.md
│   │   └── agentsail-ci/SKILL.md
│   ├── commands/agentsail/             # slash 단축
│   ├── hooks/agentsail/                # PR merge block stub만
│   ├── rules/                          # release gate 도메인 룰
│   └── settings.json
├── cmd/agentsail/main.go               # Go entrypoint
├── internal/
│   ├── cli/                            # subcommand: root, version, doctor, init,
│   │                                   #            ci, report, collect, compile,
│   │                                   #            check, verdict
│   ├── contract/                       # customer_contract.json model + compile
│   ├── evidence/                       # cache/contract/run/report 파일 IO
│   ├── adapter/                        # base + mock + http
│   └── render/                         # terminal + html(html/template) + cmux OSC9
├── pkg/version/
├── examples/
│   └── http_target/main.go             # Go HTTP mock chat endpoint (:8000)
├── .agentsail/                         # state dir (gitignore)
│   ├── cache/<customer>/{slack,gmail,notion,github}.json
│   ├── contracts/<customer>-contract.json
│   ├── runs/<customer>-run-NNN.json
│   └── reports/<customer>-run-NNN.html
├── go.mod / go.sum / Makefile
├── .gitignore                          # .agentsail/, bin/ 추가
└── docs/02-agent-sail/                 # 기존 사양
```

## 6. Phase 분해

| Phase | 단위 | 산출물 |
|---|---|---|
| **1. Scaffold** (직렬) | go.mod / Makefile / cmd 진입 / internal/cli stub / examples/http_target / .gitignore | `make build` 가 `bin/agentsail` 생성, `agentsail version` 동작, `examples/http_target` 가 `:8000` 응답 |
| **2. Plugin assets** (4 agents 병렬) | `.claude-plugin/plugin.json`, `.claude/agents/agentsail/*` (9개), `.claude/skills/agentsail*/SKILL.md` (7개), `.claude/{commands,hooks,rules}/`, `.claude/settings.json` | Claude Code 가 plugin manifest 인식, `/agentsail` 디스패처 응답 |
| **3. Go internal** (3 agents 병렬) | `internal/{contract,evidence,adapter,render}/` 본체 + `internal/cli/{compile,check,report}.go` 본체 | 각 단계 명령이 단독 호출로 fixture 입력 → fixture 출력 produce |
| **4. 통합 + e2e** (직렬) | `internal/cli/ci.go` 4단계 sequential, demo fixtures, e2e 시연 path 3개 | finbank → HOLD, retailco → SHIP, acme-bank → BLOCK 모두 HTML 자동 오픈 |

## 7. 검증 기준 (verifiable success — Karpathy 4)

- [V1] `make build` → `bin/agentsail` 생성, `bin/agentsail version` OK
- [V2] `go run ./examples/http_target` → `:8000/chat` POST 응답 OK
- [V3] `.claude-plugin/plugin.json` 로드 후 Claude Code 안에서 `/agentsail` 디스패처 응답
- [V4] fixture 가 `.agentsail/cache/finbank/*.json` 에 채워진 상태에서
       `bin/agentsail ci --customer finbank --target http://localhost:8000/chat --report --open` 실행 시
  - `.agentsail/contracts/finbank-contract.json` 생성
  - `.agentsail/runs/finbank-run-001.json` 의 `verdict == "HOLD"`
  - `.agentsail/reports/finbank-run-001.html` 생성 및 자동 오픈
- [V5] retailco 같은 명령 → `verdict == "SHIP"`
- [V6] acme-bank 같은 명령 → `verdict == "BLOCK"`, 실패 사유 3개 (`missing CSV export`, `beta badge exposed`, `tone drift`) 가 HTML 에 노출

## 8. 의존성 정책 (Karpathy 2 — Simplicity First)

- Go stdlib 우선
- 외부 의존성: `gopkg.in/yaml.v3` 만 허용
- `cobra`/`urfave/cli` 등 CLI 라이브러리 사용 **금지** — stdlib `flag` + 수동 subcommand switch
- HTML template: `html/template` (stdlib)
- HTTP client/server: `net/http` (stdlib)
- `bubbletea`/`tview` 같은 TUI 라이브러리 **금지** (TUI 라이브 차트는 다음 세션)

근거: 9시간 빌드 안에서 의존성이 늘면 build/lock/version 문제가 시연 직전에 터질 위험이 큼. stdlib + yaml.v3 면 충분히 4단계 파이프라인을 구현할 수 있다.

## 9. Surgical 영역 (Karpathy 3)

- `docs/02-agent-sail/`, `AGENTS.md`, `CLAUDE.md`, `README3.md`, `dev-list.md` 는 **건드리지 않는다**
- 새 파일은 모두 위 §5 트리 안에서 생성
- 기존 문서의 한국어/영어 톤·이모지 정책 유지

## 10. 보류 (다음 세션)

- Hotl state machine + hook 자동 step 전환
- Playwright smoke check
- 실시간 load probe (Go goroutine + ramp + p50/p95)
- rich.Live 풍 TUI 라이브 차트
- kanban board, init-survey 등 oh-my-braincrew 부수기능
- 실제 Slack/Notion/Gmail/GitHub OAuth 직접 연결 (현재는 Claude Code MCP 경유)

## 11. 시연 핵심 문장

> Same agent. Different customer. Different launch gate.

cmux 워크스페이스 3개 동시: `finbank` (HOLD 빨강 점멸) / `retailco` (SHIP 초록 점멸) / `acme-bank` (BLOCK 빨강 강조) → HTML report 3개 자동 오픈 → 발표자 한 문장으로 마감.
