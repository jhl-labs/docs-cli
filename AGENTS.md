# AGENTS.md — docs-cli 작업 규칙

> 이 저장소에서 작업하는 모든 에이전트/기여자가 지켜야 하는 규칙이다.
> `docs-cli` 는 **표준 문서 스키마(단일 출처)** 를 코드로 고정해 두고, 그 스키마에서
> 스캐폴딩·생성·검증·렌더링을 일관되게 찍어낸다.
> 핵심 한 줄: **스키마가 진실이다. 스키마를 바꾸지 않고 산출물만 바꾸지 마라.**

---

## 0. 황금 규칙 (THE GOLDEN RULE)

코드(`internal/**`)나 문서(`docs/**`)를 건드린 직후, 매번:

```bash
gofmt -l $(find . -name '*.go' -not -path './.git/*')   # 출력이 비어 있어야 함
go vet ./...
go test ./...
go run ./cmd/docs-cli validate . --strict
```

- 위 네 가지가 모두 깨끗해야 커밋한다.
- `validate` 오류(ERROR)가 0이 아니면 커밋하지 않는다.

---

## 1. 단일 출처 = `internal/schema`

표준 문서 패턴(섹션·문서·챕터·프론트매터 키)은 **오직** [`internal/schema/schema.go`](internal/schema/schema.go) 에 정의된다.

- 문서를 추가/삭제하거나 챕터를 바꾸려면 **스키마를 고친다.** 템플릿이나 검증기를 따로 손대지 않는다.
- 스캐폴더·프롬프트 빌더·검증기·렌더러는 모두 스키마를 읽는다. 스키마를 바꾸면 네 도구가 함께 따라온다.
- 스키마 구조를 바꾸면 `SchemaVersion` 을 올리고, `internal/schema/schema_test.go` 와 영향 받는 테스트를 갱신한다.

---

## 2. 도구 사이의 정합성 (깨지면 안 되는 불변식)

| 불변식 | 강제 위치 |
| --- | --- |
| 모든 문서는 프론트매터 + 스키마 챕터를 가진다 | `validate` (필수 문서는 ERROR) |
| `doc_id`/`section` 은 스키마와 일치한다 | `validate` |
| `reviewed` 문서에 `TODO(docs-cli)` 자리표시자 금지 | `validate` |
| 스캐폴드 산출물은 항상 검증을 통과한다 | `internal/validate/validate_test.go` |
| 프론트매터는 round-trip(파싱→직렬화) 안정적이다 | `internal/mddoc/mddoc_test.go` |

새 불변식을 도입하면 **산문이 아니라 테스트로** 강제한다. tiny-sso 의 추적성 규율과 동일하게, "내가 봤다"로 대체하지 않는다.

---

## 3. 의존성 규칙

- **외부 의존성 0** 을 유지한다(표준 라이브러리만). `go.mod` 에 require 를 추가하지 않는다.
- 패키지 의존 방향: `cli → (schema, scaffold, agent, validate, render, skill, project)`, 그리고 그들 → `mddoc`/`schema`. 역방향 의존 금지.

---

## 4. 커밋·언어 규칙

- 문서·주석·커밋 메시지는 한국어를 기본으로 한다(코드 식별자는 영어).
- 변경은 §0 게이트가 깨끗할 때만 커밋한다.
- 스키마를 바꾸면 [README.md](README.md) 의 스키마 표와 `make dogfood` 산출물(`docs/`, `skill.md`)을 함께 갱신한다.

---

## 5. 빠른 참조

```bash
make build          # 버전 주입 빌드
make test
make dogfood        # docs-cli 자신의 docs/ 와 skill.md 재생성 후 validate
go run ./cmd/docs-cli --help
```
