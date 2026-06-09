# docs-cli

`docs-cli` is a local-first Go tool that turns any repository into a **standardized, web-portable documentation set** — and can drive an AI agent to reverse-engineer the codebase and fill it in.

이 저장소는 `docs-cli` 의 제품/구현 기준 문서와 Go 기반 CLI 구현을 함께 포함합니다. 하나의 **표준 문서 스키마**(단일 출처)에서 `docs/` 폴더 트리·각 문서의 프론트매터·챕터 구조를 동일하게 찍어내므로, 프로젝트가 달라도 산출물의 형태가 일정합니다. 그 결과물은 그대로 웹 서비스로 이식할 수 있고, AI 에이전트가 소비하기에도 적합합니다.

## Goals

- Python, TypeScript, Go, Rust, C#, Java 프로젝트의 `docs/` 폴더를 **하나의 표준 패턴**으로 생성합니다.
- 같은 스키마로 스캐폴딩 → (에이전트) 채우기 → 검증 → 렌더링을 한 워크플로로 수행합니다.
- 산출물은 **정형화된 프론트매터 + 고정된 챕터 구조**라서 웹 서비스가 매핑 없이 그대로 ingest 할 수 있습니다.
- `claude`, `codex`, `gemini`, `opencode` 같은 AI 에이전트를 비대화형으로 호출해 리버스 엔지니어링을 자동화합니다.
- 출력 형식은 Markdown(원본), HTML(문서 사이트), XML(웹 서비스 ingest)을 지원합니다.
- 단일 실행 파일로 배포하며 Linux, macOS, Windows 를 모두 지원합니다.
- GitHub Actions 사용 시 `runs-on: jhl-space` self-hosted runner 를 기본 예시로 둡니다.

## Status

Current implementation:

- `docs-cli init [path]`: 표준 스키마 v1(17개 문서)에 따라 `docs/` 트리를 스캐폴딩합니다. 각 문서는 프론트매터와 챕터 자리표시자를 갖습니다.
- `docs-cli generate [path]`: 문서별 리버스 엔지니어링 프롬프트를 생성하고, `--agent` 지정 시 에이전트를 비대화형으로 실행해 `docs/<id>.md` 를 채웁니다. `--dry-run` 은 프롬프트 파일만 작성합니다.
- `docs-cli validate [path]`: 필수 문서 존재·프론트매터 키·챕터 누락·`reviewed` 문서의 자리표시자 잔존을 검사합니다.
- `docs-cli render [path]`: 각 문서를 HTML(문서 사이트)과 XML(구조적 표현)로 렌더링합니다.
- `docs-cli doctor`: 작업 디렉터리 쓰기 가능 여부, 사용 가능한 AI 에이전트, 현재 `docs/` 정합성을 점검합니다.
- `docs-cli skill` (= `--generate-skill`): 에이전트가 `docs-cli` 를 올바르게 쓰도록 가르치는 superpowers 형식의 `SKILL.md` 를 생성합니다.
- 언어 자동 감지(매니페스트 마커 → 소스 확장자 폴백), 결정적(deterministic) 스캐폴딩.

Not implemented yet: 양방향 동기화(렌더된 HTML→MD 역변환), 다국어 동시 출력 트리, 스키마 v2(사용자 정의 문서셋), Mermaid 사전 렌더.

## How It Works

```
                    ┌──────────────────────────────────────────┐
                    │   internal/schema  (표준 패턴 = 단일 출처)   │
                    └──────────────────────────────────────────┘
                        │            │            │           │
            ┌───────────┘     ┌──────┘      ┌─────┘      ┌────┘
            ▼                 ▼             ▼            ▼
   scaffold (init)     agent (generate)  validate     render
   docs/ 트리 생성       프롬프트/에이전트     정합성 게이트   html · xml
            │                 │             │            │
            └───── docs/*.md (frontmatter + 고정 챕터) ───┘
                                  │
                                  ▼
                       웹 서비스 ingest (doc_id·section·chapter)
```

스키마는 문서·섹션·챕터를 Go 데이터로 고정해 둔 **단일 출처**입니다. 스캐폴더·프롬프트 빌더·검증기·렌더러가 모두 같은 스키마를 읽으므로, 네 도구의 산출물이 어긋나지 않습니다. tiny-sso 의 추적성(traceability) 규율과 superpowers 의 스킬 작성 원칙을 이 구조에 반영했습니다.

## The Standardized Schema (v1)

문서는 6개 섹션으로 묶입니다. 각 문서는 동일한 프론트매터 키와, 스키마가 정한 H2 챕터를 **정확히 그 순서로** 가집니다.

| Section | 문서 (doc_id) | 필수 |
| --- | --- | --- |
| overview | `overview`, `glossary` | ✅ |
| architecture | `architecture`, `components`, `data-model`, `diagrams/` | `architecture`·`components` ✅ |
| interfaces | `api`, `cli-reference`, `config-reference` | — |
| operations | `build-and-release`, `deployment`, `development`, `testing`, `security` | `build-and-release`·`development`·`testing`·`security` ✅ |
| decisions | `adr/` (인덱스 + 개별 ADR) | ✅ |
| reference | `roadmap`, `traceability` | `traceability` ✅ |

### Frontmatter contract

모든 문서 첫 머리에 아래 블록이 들어갑니다(키는 고정, 값만 채움). 웹 서비스는 이 키만 알면 모든 문서를 동일하게 다룰 수 있습니다.

```yaml
---
doc_id: overview              # 안정적 슬러그 = 파일명 = 웹 라우트
title: 프로젝트 개요
section: overview             # 6개 섹션 중 하나
order: 1                      # 섹션 내 정렬
audience: [newcomer, decision-maker]
status: draft                 # draft | generated | reviewed
schema_version: 1
generated_by: docs-cli template   # 또는 에이전트 이름
source_commit: a1b2c3d
updated: 2026-06-09
---
```

전체 문서·챕터 목록은 `docs-cli skill` 출력 또는 [docs/SCHEMA.md](docs/SCHEMA.md) 를 참고하세요.

## Install

공개 바이너리를 `jhl-labs/dist` 에서 설치합니다(릴리스 파이프라인이 구성된 이후):

```bash
curl -fsSL https://jhl-labs.github.io/docs-cli/install.sh | sudo bash
```

버전 고정 또는 사용자 쓰기 가능 경로 설치:

```bash
curl -fsSL https://jhl-labs.github.io/docs-cli/install.sh | sudo env VERSION=v0.1.0 bash
curl -fsSL https://jhl-labs.github.io/docs-cli/install.sh | INSTALL_DIR="$HOME/.local/bin" bash
```

소스에서 빌드:

```bash
git clone git@github.com:jhl-labs/docs-cli.git
cd docs-cli
make build           # bin/docs-cli
go test ./...
go run ./cmd/docs-cli --help
```

## Usage

```bash
docs-cli init .                                  # 표준 docs/ 트리 스캐폴딩
docs-cli init . --output-dir documentation --force

docs-cli generate .                              # 프롬프트 파일만 생성 (--agent none)
docs-cli generate . --agent claude               # 에이전트로 전체 문서 채우기
docs-cli generate . --agent codex --doc architecture --doc components
docs-cli generate . --dry-run --print-prompt     # 프롬프트를 stdout 으로 미리보기

docs-cli validate .                              # 표준 정합성 검사
docs-cli validate . --strict                     # 경고도 실패 처리

docs-cli render .                                # docs/_site/*.html
docs-cli render . --format html --format xml     # html + xml
docs-cli render . --format xml --output-dir build/docs-xml

docs-cli doctor                                  # 환경·에이전트 점검
docs-cli --generate-skill --output skill.md      # 에이전트용 SKILL.md
```

### Exit Codes

| Code | Meaning |
| --- | --- |
| 0 | 성공 |
| 1 | 검증 실패 (validate 오류, 또는 strict 모드 경고) |
| 2 | 잘못된 명령·플래그·입력 |
| 3 | 에이전트 실행 실패 |
| 4 | 환경 문제 (쓰기 불가, 에이전트 없음 등) |

## Reverse Engineering with Agents

`generate` 는 문서마다 다음을 수행합니다.

1. 스키마에서 그 문서의 **필수 챕터와 작성 안내**, 그리고 **프론트매터 계약**을 담은 프롬프트를 만듭니다.
2. 프롬프트를 `.docs-cli/prompts/<id>.prompt.md` 에 저장합니다(감사·재현용).
3. `--agent` 지정 시 해당 CLI 를 비대화형으로 실행하고, 표준 출력 결과를 `docs/<id>.md` 로 기록합니다.

| Agent | 호출 형태 |
| --- | --- |
| `claude` | `claude -p <prompt>` |
| `codex` | `codex exec <prompt>` |
| `gemini` | `gemini -p <prompt>` |
| `opencode` | `opencode run <prompt>` |

에이전트가 없을 때는 `--dry-run` 으로 프롬프트만 만들고, 결과물을 직접 `docs/<id>.md` 에 붙여 넣어도 됩니다. 어느 경로든 `docs-cli validate` 가 최종 게이트입니다.

## Output Formats

| Format | 용도 | 산출 위치(기본) |
| --- | --- | --- |
| Markdown | 원본·편집·리뷰 | `docs/<id>.md` |
| HTML | 문서 사이트 (GitHub Pages) | `docs/_site/<id>.html` |
| XML | 웹 서비스 ingest (구조적: doc/section/chapter) | `docs/_xml/<id>.xml` |

HTML 변환기는 표준 문서가 쓰는 Markdown 부분집합(헤딩·펜스 코드·목록·표·인용·링크·강조·인라인 코드)을 처리하며, 프론트매터는 `<meta>` 로 옮기고 본문에서 제거합니다. XML 변환기는 H2 챕터를 `<chapter id heading><markdown>…</markdown>` 로 구조화합니다.

## GitHub Actions

`docs-cli` 를 CI 에 연결해 문서 정합성을 PR 게이트로 두고, 산출물을 아티팩트로 올립니다. self-hosted runner `jhl-space` 를 기본 예시로 사용합니다.

```yaml
name: docs

on:
  pull_request:
  push:
    branches: [main, master]

jobs:
  docs:
    runs-on: jhl-space
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true
      - run: go build -o bin/docs-cli ./cmd/docs-cli
      - name: Validate docs
        run: bin/docs-cli validate . --strict
      - name: Render docs
        run: bin/docs-cli render . --format html --format xml
      - uses: actions/upload-artifact@v7
        if: always()
        with:
          name: docs-site
          path: docs/_site
```

재사용 가능한 composite action 은 별도 저장소 [`jhl-labs/docs-cli-action`](https://github.com/jhl-labs/docs-cli-action) 으로 제공됩니다. 검증·렌더는 물론 AI 에이전트로 문서를 채우는 단계까지 자동화할 수 있습니다:

```yaml
- uses: jhl-labs/docs-cli-action@main
  with:
    command: generate
    agent: claude
    target: .
```

자세한 내용은 [CI/CD 가이드](guides/ci-cd.md) 를 참고하세요.

## Versioning and Release

- **SemVer** 태그(`vMAJOR.MINOR.PATCH`)로 릴리스합니다. 버전·커밋·날짜는 빌드 시 `-ldflags` 로 `internal/version` 에 주입합니다.
- `make build` 가 로컬 빌드를, [`scripts/build-release-artifacts.sh`](scripts/build-release-artifacts.sh) 가 6개 플랫폼 크로스 빌드 + `SHA256SUMS` 를 만듭니다.
- 태그 push(`v*`) 시 [release 워크플로](.github/workflows/release-dist.yml)가 산출물을 빌드하고 공개 `jhl-labs/dist` 저장소로 게시합니다.
- `docs/**` 변경 시 [pages 워크플로](.github/workflows/pages.yml)가 GitHub Pages 로 문서 사이트를 배포합니다.

자세한 절차는 docs 의 [build-and-release](docs/build-and-release.md) 와 [Release 가이드](guides/dist-release.md) 를 참고하세요.

## Project Layout

```text
cmd/docs-cli/          CLI entry point
internal/version/      버전 정보 (-ldflags 주입)
internal/schema/       표준 문서 패턴 — 단일 출처
internal/mddoc/        프론트매터 + 헤딩 파서/직렬화 (의존성 없음)
internal/project/      구현 언어 감지 (py/ts/go/rust/c#/java)
internal/scaffold/     init — 스키마 → docs/ 트리 생성
internal/agent/        에이전트 어댑터 + 리버스 엔지니어링 프롬프트
internal/validate/     표준 스키마 정합성 검사
internal/render/       Markdown → HTML / XML
internal/skill/        SKILL.md 생성기 (superpowers 형식)
internal/cli/          명령 디스패치·플래그·도움말
guides/                사용자/운영자 가이드
site/                  GitHub Pages 랜딩 사이트 + install.sh
docs/                  docs-cli 자신의 표준 문서 (dogfooding)
```

## Guides

- [Toolchain Guide](guides/toolchain.md): 빌드 도구·에이전트·환경.
- [CI/CD Guide](guides/ci-cd.md): GitHub Actions·문서 게이트·페이지 배포.
- [Release Guide](guides/dist-release.md): 버전·태그·공개 배포 절차.

## Design Influences

- **tiny-sso/docs** — 추적성(traceability) 규율, 섹션화된 문서 지도, "정의했으면 끝까지 잇는다" 원칙. `validate` 는 이 규율을 기계가 판정하게 만든 것입니다.
- **superpowers** — 스킬 작성 원칙(트리거 중심 description, Iron Law, 합리화 차단). `docs-cli skill` 산출물이 이 형식을 따릅니다.
- **security-cli** — 단일 바이너리·`internal/` 레이아웃·수동 플래그 파싱·버전 주입·릴리스/페이지 워크플로·`jhl-space` runner 관례.

## License

License is not finalized yet. Until a `LICENSE` file is added, treat this repository as private/internal material owned by JHL Labs.
