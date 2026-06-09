# CI/CD Guide

`docs-cli` 를 GitHub Actions 에 연결해 문서 정합성을 게이트로 두고, 산출물을 배포하는 방법입니다. self-hosted runner `jhl-space` 를 기본 예시로 사용합니다.

## Pull Request 문서 게이트

PR 에서는 표준 정합성을 빠르게 검증합니다.

```yaml
name: docs

on:
  pull_request:

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
      - run: bin/docs-cli validate . --strict
      - run: bin/docs-cli render . --format html --format xml
      - uses: actions/upload-artifact@v7
        if: always()
        with:
          name: docs-site
          path: docs/_site
```

## 재사용 가능한 Action

이 저장소의 [`action.yaml`](../action.yaml) 을 그대로 호출할 수 있습니다.

```yaml
- uses: jhl-labs/docs-cli@main
  with:
    command: validate
    target: .
    strict: "true"
```

에이전트로 문서를 채우는 단계까지 포함하려면:

```yaml
- uses: jhl-labs/docs-cli@main
  with:
    command: generate
    agent: claude
    target: .
    validate: "true"
```

## 문서 사이트 배포 (GitHub Pages)

`docs/**` 가 바뀌면 [pages 워크플로](../.github/workflows/pages.yml)가 문서를 HTML 로 렌더링하고 `install.sh` 와 함께 GitHub Pages 로 배포합니다. 설치 스크립트는 `https://jhl-labs.github.io/docs-cli/install.sh` 에서 제공됩니다.

## 보관할 산출물

- 문서 사이트: `docs/_site/**`
- 웹 서비스 ingest 용: `docs/_xml/**`
- 검증 로그: `docs-cli validate` 출력

## 정책 권장값

- PR 은 `validate --strict` 로 경고도 실패 처리합니다.
- `generate` 로 채운 문서는 사람이 검토 후 `status: reviewed` 로 올립니다(검토 전에는 `generated` 유지).
- CI 는 실패해도 산출물을 업로드합니다(`if: always()`).
