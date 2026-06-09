# Toolchain Guide

`docs-cli` 를 빌드·실행하는 데 필요한 도구와 환경을 정리합니다.

## 빌드 도구

| 도구 | 용도 | 비고 |
| --- | --- | --- |
| Go 1.26+ | 빌드·테스트 | `go.mod` 의 `go` 지시문 기준 |
| make | 태스크 러너 | 선택. `make build`, `make test`, `make dogfood` |
| git | 버전·커밋 메타데이터 | `init` 의 `source_commit` 채움에 사용 |

`docs-cli` 는 표준 라이브러리만 사용합니다(외부 의존성 0). 따라서 인터넷 없이도 `go build`/`go test` 가 가능합니다.

## AI 에이전트 (선택)

`generate` 가 문서를 자동으로 채우려면 아래 중 하나의 비대화형 CLI 가 PATH 에 있어야 합니다.

| 에이전트 | 호출 형태 | 확인 |
| --- | --- | --- |
| Claude Code | `claude -p <prompt>` | `docs-cli doctor` |
| Codex | `codex exec <prompt>` | `docs-cli doctor` |
| Gemini CLI | `gemini -p <prompt>` | `docs-cli doctor` |
| opencode | `opencode run <prompt>` | `docs-cli doctor` |

에이전트가 없으면 `docs-cli generate --dry-run` 으로 `.docs-cli/prompts/` 에 프롬프트만 생성하고, 임의의 에이전트/사람이 채워 `docs/<id>.md` 에 기록할 수 있습니다.

## 환경 점검

```bash
docs-cli doctor
```

- 작업 디렉터리 쓰기 가능 여부
- 설치된 AI 에이전트 목록
- 현재 `docs/` 의 검증 오류/경고 수
