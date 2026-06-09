# Release Guide

`docs-cli` 의 버전 관리와 공개 배포 절차입니다.

## 버전 관리

- **SemVer** 태그를 사용합니다: `vMAJOR.MINOR.PATCH` (예: `v0.1.0`, 프리릴리스 `v0.1.0-rc.1`).
- 버전·커밋·날짜는 빌드 시 `-ldflags` 로 `internal/version` 패키지에 주입됩니다.

```bash
make build VERSION=v0.1.0          # bin/docs-cli, 버전 주입됨
./bin/docs-cli --version
```

## 릴리스 산출물

[`scripts/build-release-artifacts.sh`](../scripts/build-release-artifacts.sh) 가 6개 플랫폼을 크로스 빌드하고 `SHA256SUMS` 를 만듭니다.

```bash
VERSION=v0.1.0 OUT_DIR=dist/release scripts/build-release-artifacts.sh
```

대상 플랫폼: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64`. 각 플랫폼은 직접 실행 바이너리와 `.tar.gz` 아카이브로 제공됩니다.

## 공개 배포

태그를 push 하면 [release 워크플로](../.github/workflows/release-dist.yml)가 다음을 수행합니다.

1. 포맷·모듈·vet·테스트 검사.
2. 릴리스 산출물 빌드 + `SHA256SUMS`.
3. 릴리스 노트 생성(직전 태그 이후 커밋 로그).
4. 공개 `jhl-labs/dist` 저장소에 `docs-cli-<version>` 태그로 GitHub Release 게시(`DIST_REPO_TOKEN` 필요).

```bash
git tag v0.1.0
git push origin v0.1.0
```

또는 `workflow_dispatch` 로 버전을 직접 지정해 수동 실행할 수 있습니다.

## 설치 경로

배포 후 사용자는 GitHub Pages 의 설치 스크립트로 받습니다.

```bash
curl -fsSL https://jhl-labs.github.io/docs-cli/install.sh | sudo bash
curl -fsSL https://jhl-labs.github.io/docs-cli/install.sh | sudo env VERSION=v0.1.0 bash
```

설치 스크립트는 `jhl-labs/dist` 의 최신 `docs-cli-v*` 태그를 찾아 현재 OS/아키텍처용 바이너리를 내려받고 `SHA256SUMS` 로 검증합니다.

## 필요한 시크릿

| 시크릿 | 용도 |
| --- | --- |
| `DIST_REPO_TOKEN` | 공개 `jhl-labs/dist` 저장소에 릴리스를 게시할 토큰 |
