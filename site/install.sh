#!/usr/bin/env bash
# docs-cli installer.
#
# Usage:
#   curl -fsSL https://jhl-labs.github.io/docs-cli/install.sh | sudo bash
#   curl -fsSL https://jhl-labs.github.io/docs-cli/install.sh | sudo env VERSION=v0.1.0 bash
#   curl -fsSL https://jhl-labs.github.io/docs-cli/install.sh | INSTALL_DIR="$HOME/.local/bin" bash
set -euo pipefail

APP_NAME="docs-cli"
DIST_REPO="${DIST_REPO:-jhl-labs/dist}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

err() { echo "error: $*" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

have curl || err "curl is required"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${os}" in
  linux|darwin) ;;
  *) err "unsupported OS: ${os} (use the .exe asset on Windows)";;
esac

arch="$(uname -m)"
case "${arch}" in
  x86_64|amd64) arch="amd64";;
  aarch64|arm64) arch="arm64";;
  *) err "unsupported architecture: ${arch}";;
esac

api="https://api.github.com/repos/${DIST_REPO}/releases"
auth=()
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
  auth=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
elif [[ -n "${GH_TOKEN:-}" ]]; then
  auth=(-H "Authorization: Bearer ${GH_TOKEN}")
fi

if [[ "${VERSION}" == "latest" ]]; then
  # Resolve the newest docs-cli-v* tag from the dist repo.
  tag="$(curl -fsSL "${auth[@]}" "${api}?per_page=100" \
    | grep -o "\"tag_name\": *\"${APP_NAME}-v[^\"]*\"" \
    | head -n1 | sed -E 's/.*"(.*)"/\1/')"
  [[ -n "${tag}" ]] || err "could not resolve latest ${APP_NAME} release in ${DIST_REPO}"
else
  tag="${APP_NAME}-${VERSION}"
fi

version="${tag#${APP_NAME}-}"
base="https://github.com/${DIST_REPO}/releases/download/${tag}"
asset="${APP_NAME}_${version}_${os}_${arch}"

tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT

echo "downloading ${asset} (${tag})"
curl -fsSL -o "${tmp}/${APP_NAME}" "${base}/${asset}" || err "failed to download ${asset}"

if curl -fsSL -o "${tmp}/SHA256SUMS" "${base}/SHA256SUMS" 2>/dev/null; then
  if have sha256sum; then
    expected="$(grep " ${asset}\$" "${tmp}/SHA256SUMS" | awk '{print $1}')"
    actual="$(sha256sum "${tmp}/${APP_NAME}" | awk '{print $1}')"
    if [[ -z "${expected}" ]]; then
      err "no checksum for ${asset} in SHA256SUMS"
    fi
    if [[ "${expected}" != "${actual}" ]]; then
      err "checksum mismatch for ${asset}"
    fi
    echo "checksum verified"
  fi
fi

chmod +x "${tmp}/${APP_NAME}"
mkdir -p "${INSTALL_DIR}"
mv "${tmp}/${APP_NAME}" "${INSTALL_DIR}/${APP_NAME}" \
  || err "failed to install to ${INSTALL_DIR} (try sudo, or set INSTALL_DIR)"

echo "installed ${APP_NAME} ${version} to ${INSTALL_DIR}/${APP_NAME}"
"${INSTALL_DIR}/${APP_NAME}" --version || true
