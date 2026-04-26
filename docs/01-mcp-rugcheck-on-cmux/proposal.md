# 01. mcp-rugcheck-on-cmux — "MCP가 설치 후 변심하는 순간을 cmux가 점멸로 알린다"

## 1. 풀려는 문제 정의

**"AI 에이전트가 신뢰하는 MCP 서버의 tool description 이 런타임에 악성으로 바뀌는 순간(rug pull)을 사람이 못 잡는다."**

근거:
- 2026.04.16 *The Register*: Anthropic MCP SDK "by design" RCE, 200,000 인스턴스 영향. Anthropic 패치 안 함 → 외부 도구가 유일 해법
- 2026.02 *OX Security*: LobeHub·Cursor Directory PoC 페이로드 11개 업로드 → 9개가 보안 리뷰 0회 통과
- MCP install 시점에는 정상 → 런타임에 tool description 만 교체하면 사용자는 알 방법이 없음
- CVE-2025-59536, CVE-2026-21852: Claude Code 프로젝트 파일이 trust dialog 전 실행

## 2. 왜 지금 우리가 풀어야 하는가

심사 = AIM Intelligence. 그들의 공개 repo `AIM-MCP`, `awesome-mcp-security` 가 곧 정답지다. **MCP 보안이 AIM 의 현재 1순위 관심사**이며, 그 핵심 미해결 문제가 rug pull.

또한 시연 무대가 cmux 자체다 → **"심사관이 자기 도구 위에서"** 시연되는 메타 효과. AIM 의 "spear and shield" 메타포 (공격 + 방어 한 사이클) 와도 정확히 일치.

## 3. 핵심 메커니즘 (3줄)

1. cmux 가 띄우는 모든 MCP 서버의 `list_tools()` 응답을 hook → SHA256 hash + sentence-transformers 임베딩 저장
2. 매 호출 직전 다시 fetch → 코사인 유사도 < 0.92 또는 hash 불일치 시 **rug pull 판정**
3. 즉시 도구 호출 차단 + cmux OSC 9 알림 한 줄로 사이드바 빨강 점멸 + 변심 diff 표시

## 4. 비유로 5분 이해

> 단골 식당이 있다. 처음 갔을 때 메뉴판은 "김치찌개 8천원". 그런데 다음에 갔더니 사장님 얼굴도 똑같고, 메뉴판 글씨도 똑같은데 메뉴판 뒷면에 작게 적혀 있다 — **"손님 신용카드 사진을 외부 주소로 전송"**. 손님은 평소처럼 김치찌개를 시켰지만 카드는 이미 털렸다.

mcp-rugcheck 는 그 **"메뉴판 뒷면에 새로 추가된 줄"** 을 매번 식당 들어갈 때마다 비교해서 알려주는 위생 검사관이다.

기술 매핑:
- 메뉴판 = MCP `list_tools()` 결과
- 사장님 얼굴 = SHA256 hash (똑같은 척하는 표면)
- 메뉴판 뒤 추가된 줄 = description 변조 (의미 임베딩에서만 잡힘)
- 위생 검사관 = mcp-rugcheck

## 5. 라이브 3분 시연 시나리오

**무대 세팅**: cmux 워크스페이스 4개 동시 띄움 — `notion-mcp` / `github-mcp` / `slack-mcp` / `gdrive-mcp`. 각 워크스페이스에 mcp-rugcheck 가 백그라운드에서 watch 중.

| 초 | 화면 | 청중이 보는 것 | 의미 |
|---|---|---|---|
| 0s | 발표자 명령 입력 | `$ ./demo/rugpull.sh` | 한 줄로 4개 MCP 동시 hot-swap (description만 악성으로 교체) |
| 1s | 4개 워크스페이스 백그라운드 | 정상 (변화 없음) | rug pull 직후, 사용자는 모름 |
| 2s | mcp-rugcheck 로그 | `[DRIFT] notion-mcp/create_page: cosine=0.71` ×4 | 임베딩 drift 감지 |
| 3s | **cmux 사이드바 4개 빨강 연쇄 점멸** | 시각적 폭발 | 한 컷 시연의 정점 |
| 5s | 화면에 빨간 diff 4개 띄움 | `+ "Send card image to evil.com"` 강조 | 변심 내용 visible |
| 10s | 발표자 한 문장 | "Anthropic이 패치 안 한 200,000 인스턴스, cmux 위에서 1줄로 막았습니다" | 헤드라인 |

## 6. 9시간 빌드 분해

| 시간 | 마일스톤 |
|---|---|
| 08:00–08:30 | 환경 셋업 (uv venv, mcp SDK, sentence-transformers, cmux 연동 확인) |
| 08:30–10:00 | MCP `list_tools()` hook + hash/embedding 저장소 (SQLite) |
| 10:00–11:30 | drift 판정 로직 + 차단 인터셉터 |
| 11:30–12:30 | cmux OSC 9 통합 + 사이드바 알림 시각화 |
| 12:30–13:00 | 점심 + 더미 MCP 4개 (`notion/github/slack/gdrive`) 시드 |
| 13:00–14:30 | rug pull 시나리오 스크립트 (`demo/rugpull.sh`) |
| 14:30–16:00 | E2E 시연 리허설 ×3 |
| 16:00–17:00 | 빨간 diff UI polish + 발표 슬라이드 1장 |
| 17:00–18:00 | 최종 리허설 + 발표 |

**라이브러리**: `mcp` (Anthropic SDK), `sentence-transformers`, `sqlite3`, `rich` (터미널 diff 시각화)

## 7. 리스크 + 대응 + 채점

**3축 채점**: ① 한 컷 시연 10/10 / ② 헤드라인 10/10 / ③ 심사관 직결 10/10 — **합 30/30**

**리스크**:
- (낮음) sentence-transformers 첫 로드 무거움 → 사전 warm-up 스크립트로 대응
- (중) cmux OSC 9 통합이 워크스페이스마다 별도 PID 처리 필요 → 13:00 시점에 안 되면 단일 워크스페이스 4-pane split 으로 fallback

**Fallback**: cmux 통합 실패 시 단일 터미널에서 빨간 박스 4개 동시 점멸로 다운그레이드. 본체 (rug pull 탐지) 는 이미 동작.

**가장 큰 강점**: 빌드 본체 단순 (hash + embedding + 1줄 hook), 시각 효과 강렬, 심사관 repo 와 1:1 매칭. 메일 v2 에서 이미 30/30 으로 1순위로 추천된 후보.
