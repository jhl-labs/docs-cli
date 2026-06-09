#!/usr/bin/env bash
set -euo pipefail

APP_NAME="${APP_NAME:-docs-cli}"
VERSION="${VERSION:-dev}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)}"
DATE="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
OUT_DIR="${OUT_DIR:-dist/release}"
PLATFORMS="${PLATFORMS:-linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64}"
VERSION_PACKAGE="${VERSION_PACKAGE:-github.com/jhl-labs/docs-cli/internal/version}"
INCLUDE_DIRECT_BINARIES="${INCLUDE_DIRECT_BINARIES:-true}"
INCLUDE_ARCHIVES="${INCLUDE_ARCHIVES:-true}"

rm -rf "${OUT_DIR}"
mkdir -p "${OUT_DIR}"

ldflags="-s -w"
ldflags="${ldflags} -X ${VERSION_PACKAGE}.Version=${VERSION}"
ldflags="${ldflags} -X ${VERSION_PACKAGE}.Commit=${COMMIT}"
ldflags="${ldflags} -X ${VERSION_PACKAGE}.Date=${DATE}"

for platform in ${PLATFORMS}; do
  goos="${platform%/*}"
  goarch="${platform#*/}"
  ext=""
  if [[ "${goos}" == "windows" ]]; then
    ext=".exe"
  fi

  package_dir="${APP_NAME}_${VERSION}_${goos}_${goarch}"
  asset_name="${package_dir}${ext}"
  work_dir="$(mktemp -d)"
  target_dir="${work_dir}/${package_dir}"
  binary_path="${target_dir}/${APP_NAME}${ext}"
  mkdir -p "${target_dir}"

  echo "building ${APP_NAME} ${VERSION} for ${goos}/${goarch}"
  CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" go build -trimpath \
    -ldflags "${ldflags}" \
    -o "${binary_path}" ./cmd/docs-cli

  if [[ "${INCLUDE_DIRECT_BINARIES}" == "true" ]]; then
    cp "${binary_path}" "${OUT_DIR}/${asset_name}"
    chmod 0755 "${OUT_DIR}/${asset_name}" 2>/dev/null || true
  fi

  if [[ "${INCLUDE_ARCHIVES}" == "true" ]]; then
    tar -C "${target_dir}" -czf "${OUT_DIR}/${package_dir}.tar.gz" .
  fi

  rm -rf "${work_dir}"
done

(
  cd "${OUT_DIR}"
  mapfile -d '' artifacts < <(find . -maxdepth 1 -type f ! -name SHA256SUMS -printf '%P\0' | sort -z)
  if [[ "${#artifacts[@]}" -eq 0 ]]; then
    echo "no release artifacts were produced" >&2
    exit 1
  fi
  sha256sum "${artifacts[@]}" > SHA256SUMS
)

echo "wrote artifacts to ${OUT_DIR}"
