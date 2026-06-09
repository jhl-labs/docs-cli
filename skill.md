---
name: standardizing-project-docs
description: Use when a project needs a standardized docs/ folder, when reverse-engineering an existing codebase into architecture docs, or when generating docs that a web service will ingest
---

# Standardizing Project Docs

## Overview

`docs-cli` produces a **fixed, web-portable documentation set** from one schema. Every document carries the same frontmatter keys and the same ordered chapters, so the output is uniform across projects and ingestible by a downstream service.

**Core principle:** The schema is the single source of truth. Never invent documents or chapters — scaffold from the schema, fill the chapters, then validate.

## When to Use

- A repository has no docs/, or has ad-hoc docs that need standardizing.
- You must reverse-engineer an unfamiliar codebase into architecture docs.
- Generated docs must be consumed by a web service (consistent structure required).
- Supported project languages: Python, TypeScript, Go, Rust, C#, Java.

**Do NOT use for:** one-off READMEs, or free-form design notes that need no structure.

## The Iron Law

```
NO DOCUMENT MARKED reviewed WITHOUT PASSING `docs-cli validate`
```

Filled a doc? Run validate. Errors > 0? Fix before committing. Leaving `TODO(docs-cli)` placeholders in a `reviewed` doc is a validation failure, not a style nit.

## Workflow

```bash
docs-cli init .                 # 1. scaffold the standardized tree into docs/
docs-cli generate . --agent claude   # 2. let an agent fill each chapter from the code
docs-cli validate .             # 3. gate: required docs, frontmatter, chapters
docs-cli render . --format html --format xml   # 4. emit web-portable artifacts
```

Without an agent, `generate --dry-run` writes one prompt file per document under `.docs-cli/prompts/`; hand those to any agent and write the result back to the matching `docs/<id>.md`.

## Quick Reference

| Command | Purpose |
| --- | --- |
| `docs-cli init [path]` | Scaffold docs/ from the schema |
| `docs-cli generate [path]` | Build prompts / run an agent to fill docs |
| `docs-cli validate [path]` | Check conformance to the schema |
| `docs-cli render [path]` | Convert docs to HTML/XML |
| `docs-cli doctor` | Check environment and available agents |
| `docs-cli skill` | Regenerate this skill |

## The Standardized Document Set (schema v1)

Fill every chapter; do not add or drop documents.

| Document | Section | Chapters |
| --- | --- | --- |
| `overview` *(required)* | overview | TL;DR · 문제 정의 · 사용자와 이해관계자 · 핵심 가치와 비-목표 · 기술 스택 요약 |
| `glossary` *(required)* | overview | 도메인 용어 · 기술 용어 |
| `architecture` *(required)* | architecture | 시스템 컨텍스트 · 컨테이너 / 런타임 단위 · 품질 속성 · 횡단 관심사 |
| `components` *(required)* | architecture | 모듈 지도 · 의존 규칙 · 경계와 확장점 |
| `data-model` | architecture | 엔티티 · 관계 (ER) · 수명주기·정합성 |
| `diagrams` | architecture | 다이어그램 목록 |
| `api` | interfaces | 인터페이스 표면 · 요청·응답 계약 · 호환성·버저닝 |
| `cli-reference` | interfaces | 명령 카탈로그 · 전역 플래그 · 종료 코드 |
| `config-reference` | interfaces | 설정 탐색·우선순위 · 설정 키 · 예시 구성 |
| `build-and-release` *(required)* | operations | 빌드 · 버전 관리 · 릴리스 파이프라인 |
| `deployment` | operations | 배포 대상 · 배포 절차 · 롤백·복구 |
| `development` *(required)* | operations | 개발 환경 준비 · 작업 루프 · 코드 규약 |
| `testing` *(required)* | operations | 테스트 계층 · 실행 방법 · 커버리지·품질 게이트 |
| `security` *(required)* | operations | 위협 표면 · 비밀·자격증명 관리 · 공급망·의존성 |
| `adr` *(required)* | decisions | 결정 목록 · ADR 작성 규약 |
| `roadmap` | reference | 마일스톤 · 현재 집중 |
| `traceability` *(required)* | reference | 커버리지 매트릭스 · 열린 항목 등록부 |

## Frontmatter Contract

Every document begins with this block (values filled, keys unchanged):

```yaml
---
doc_id: <id>
title: <title>
section: <section>
order: <n>
audience: [<roles>]
status: draft | generated | reviewed
schema_version: 1
generated_by: <template | agent name>
source_commit: <git-sha>
updated: <YYYY-MM-DD>
---
```

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| Renaming or reordering chapters | Keep the schema's headings exactly; validate enforces them |
| Leaving `TODO(docs-cli)` in a reviewed doc | Fill the chapter, or keep status at `generated` |
| Inventing extra documents | The web service only knows schema doc_ids; add via a schema change |
| Marking `reviewed` before running validate | Run `docs-cli validate` first; 0 errors required |

## Red Flags - STOP

- Writing prose without reading the actual source first
- Committing docs with validate errors
- "The chapter doesn't apply" → write why it's N/A, don't delete the heading
