---
doc_id: cli-reference
title: CLI 레퍼런스
section: interfaces
order: 2
audience: [operator, integrator]
status: reviewed
schema_version: 1
generated_by: docs-cli template
source_commit: unknown
updated: 2026-06-09
---

# CLI 레퍼런스

> 명령·플래그·종료 코드
>
> _이 문서는 `docs-cli` 표준 스키마 v1을 따릅니다._

## 명령 카탈로그

| 명령 | 목적 |
| --- | --- |
| `docs-cli init [path]` | 표준 스키마로 `docs/` 트리를 스캐폴딩 |
| `docs-cli generate [path]` | 리버스 엔지니어링 프롬프트 생성 / 에이전트로 문서 채우기 |
| `docs-cli validate [path]` | `docs/` 가 표준 스키마를 따르는지 검증 |
| `docs-cli render [path]` | 문서를 HTML/XML 로 렌더링 |
| `docs-cli doctor` | 환경·에이전트·문서 정합성 점검 |
| `docs-cli skill` | 에이전트용 `SKILL.md` 생성 (= `--generate-skill`) |
| `docs-cli version` | 버전 출력 |

### 주요 서브 플래그

| 명령 | 플래그 | 기본값 | 설명 |
| --- | --- | --- | --- |
| `init` | `--output-dir` | `docs` | 문서 루트 |
| `init` | `--lang` | `auto` | 주 언어 강제 지정 |
| `init` | `--force` | false | 기존 파일 덮어쓰기 |
| `generate` | `--agent` | `none` | claude\|codex\|gemini\|opencode\|none |
| `generate` | `--doc` | (전체) | 특정 문서만 (반복 가능) |
| `generate` | `--dry-run` | false | 프롬프트만 작성 |
| `generate` | `--print-prompt` | false | 프롬프트를 stdout 출력 |
| `validate` | `--strict` | false | 경고도 실패 처리 |
| `render` | `--format` | `html` | html\|xml (반복 가능) |
| `render` | `--output-dir` | (포맷별) | 산출물 루트 |

## 전역 플래그

| 플래그 | 설명 |
| --- | --- |
| `-h`, `--help` | 도움말 (각 서브명령에도 적용) |
| `-v`, `--version` | 버전 출력 |
| `--generate-skill` | `skill` 명령의 최상위 별칭 |

플래그는 `--flag value` 와 `--flag=value` 두 형태를 모두 받는다. `path` 위치 인자는 분석 대상 프로젝트 경로(기본 `.`)다.

## 종료 코드

| 코드 | 의미 |
| --- | --- |
| 0 | 성공 |
| 1 | 검증 실패 (validate 오류, 또는 strict 모드 경고) |
| 2 | 잘못된 명령·플래그·입력 |
| 3 | 에이전트 실행 실패 |
| 4 | 환경 문제 (쓰기 불가, 에이전트 없음 등) |
