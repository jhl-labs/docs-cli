# 표준 문서 스키마 (Schema v1)

> 이 문서는 `docs-cli` 가 생성·검증하는 **표준 문서 패턴**의 사람이 읽는 사본입니다.
> 기계가 읽는 단일 출처는 [`internal/schema/schema.go`](../internal/schema/schema.go) 이며, 둘은 항상 일치해야 합니다.

## 프론트매터 키 (모든 문서 공통)

| 키 | 타입 | 설명 |
| --- | --- | --- |
| `doc_id` | string | 안정적 슬러그. 파일명(`<doc_id>.md`)이자 웹 라우트 |
| `title` | string | 문서 H1 제목 |
| `section` | string | 6개 섹션 중 하나 |
| `order` | int | 섹션 내 정렬 순서 |
| `audience` | list | 주 독자 역할 |
| `status` | enum | `draft` \| `generated` \| `reviewed` |
| `schema_version` | string | 생성 시 스키마 버전 |
| `generated_by` | string | `docs-cli template` 또는 에이전트 이름 |
| `source_commit` | string | 생성 시점의 git 커밋 |
| `updated` | date | 최종 갱신일(YYYY-MM-DD) |

## 섹션

| order | section | 설명 |
| --- | --- | --- |
| 1 | overview | 이 프로젝트가 무엇이고 왜 존재하는가 |
| 2 | architecture | 구조·모듈·데이터가 어떻게 짜여 있는가 |
| 3 | interfaces | 구현이 그대로 따르는 인터페이스 — API·CLI·설정 |
| 4 | operations | 어떻게 빌드·배포·테스트·보안 점검하는가 |
| 5 | decisions | 왜 이렇게 설계했는가 — 아키텍처 결정 기록 |
| 6 | reference | 로드맵·추적성 등 횡단 참고 자료 |

## 문서와 챕터

각 문서는 아래 H2 챕터를 **정확히 이 제목과 순서로** 가집니다. `validate` 가 누락을 ERROR 로 처리합니다.

| doc_id | section | 필수 | 챕터 |
| --- | --- | --- | --- |
| `overview` | overview | ✅ | TL;DR · 문제 정의 · 사용자와 이해관계자 · 핵심 가치와 비-목표 · 기술 스택 요약 |
| `glossary` | overview | ✅ | 도메인 용어 · 기술 용어 |
| `architecture` | architecture | ✅ | 시스템 컨텍스트 · 컨테이너 / 런타임 단위 · 품질 속성 · 횡단 관심사 |
| `components` | architecture | ✅ | 모듈 지도 · 의존 규칙 · 경계와 확장점 |
| `data-model` | architecture | — | 엔티티 · 관계 (ER) · 수명주기·정합성 |
| `diagrams/` | architecture | — | 다이어그램 목록 |
| `api` | interfaces | — | 인터페이스 표면 · 요청·응답 계약 · 호환성·버저닝 |
| `cli-reference` | interfaces | — | 명령 카탈로그 · 전역 플래그 · 종료 코드 |
| `config-reference` | interfaces | — | 설정 탐색·우선순위 · 설정 키 · 예시 구성 |
| `build-and-release` | operations | ✅ | 빌드 · 버전 관리 · 릴리스 파이프라인 |
| `deployment` | operations | — | 배포 대상 · 배포 절차 · 롤백·복구 |
| `development` | operations | ✅ | 개발 환경 준비 · 작업 루프 · 코드 규약 |
| `testing` | operations | ✅ | 테스트 계층 · 실행 방법 · 커버리지·품질 게이트 |
| `security` | operations | ✅ | 위협 표면 · 비밀·자격증명 관리 · 공급망·의존성 |
| `adr/` | decisions | ✅ | 결정 목록 · ADR 작성 규약 |
| `roadmap` | reference | — | 마일스톤 · 현재 집중 |
| `traceability` | reference | ✅ | 커버리지 매트릭스 · 열린 항목 등록부 |

## 검증 규칙 (`docs-cli validate`)

- 필수 문서 누락 → ERROR (선택 문서 누락 → WARN)
- 프론트매터 블록 없음 → ERROR
- `doc_id`/`section` 이 스키마와 불일치 → ERROR
- 유효하지 않은 `status` → ERROR
- 스키마 챕터 누락 → ERROR
- `schema_version` 이 현재 버전과 다름 → WARN
- `status: reviewed` 인데 `TODO(docs-cli)` 자리표시자 잔존 → ERROR
