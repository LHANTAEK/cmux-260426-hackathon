#!/usr/bin/env bash
# install.sh — Download Agent Sail from GitHub and install the agentsail CLI.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/LHANTAEK/cmux-260426-hackathon/mvp/scripts/install.sh | bash
set -euo pipefail

REPO="${AGENTSAIL_REPO:-LHANTAEK/cmux-260426-hackathon}"
REF="${AGENTSAIL_REF:-mvp}"
INSTALL_DIR="${AGENTSAIL_INSTALL_DIR:-${HOME}/.local/bin}"
BINARY_NAME="agentsail"

TMP_DIR=""
cleanup() {
  if [[ -n "${TMP_DIR}" && -d "${TMP_DIR}" ]]; then
    rm -rf "${TMP_DIR}"
  fi
}
trap cleanup EXIT

error() {
  echo "ERROR: $*" >&2
  exit 1
}

info() {
  echo "==> $*"
}

need() {
  command -v "$1" >/dev/null 2>&1 || error "'$1' is required but was not found"
}

detect_platform() {
  case "$(uname -s)" in
    Darwin|Linux) ;;
    *) error "Unsupported OS: $(uname -s). Only macOS and Linux are supported." ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64|arm64|aarch64) ;;
    *) error "Unsupported architecture: $(uname -m). Only amd64 and arm64 are supported." ;;
  esac
}

main() {
  detect_platform
  need curl
  need tar
  need go

  TMP_DIR="$(mktemp -d)"
  local archive="${TMP_DIR}/agentsail.tar.gz"
  local url="https://github.com/${REPO}/archive/${REF}.tar.gz"

  info "Downloading Agent Sail from ${REPO}@${REF}"
  curl -fsSL --progress-bar -o "${archive}" "${url}" \
    || error "Download failed: ${url}"

  info "Extracting source"
  tar -xzf "${archive}" -C "${TMP_DIR}"
  local src_dir
  src_dir="$(find "${TMP_DIR}" -mindepth 1 -maxdepth 1 -type d | head -1)"
  [[ -n "${src_dir}" ]] || error "Could not find extracted source directory"

  info "Building ${BINARY_NAME}"
  mkdir -p "${INSTALL_DIR}"
  (
    cd "${src_dir}"
    GOCACHE="${TMP_DIR}/gocache" GOMODCACHE="${TMP_DIR}/gomodcache" \
      go build -o "${INSTALL_DIR}/${BINARY_NAME}" ./cmd/agentsail
  )
  chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

  info "Installed ${INSTALL_DIR}/${BINARY_NAME}"
  "${INSTALL_DIR}/${BINARY_NAME}" version

  if [[ ":${PATH}:" != *":${INSTALL_DIR}:"* ]]; then
    echo ""
    echo "WARNING: ${INSTALL_DIR} is not in PATH."
    echo "Add this to your shell profile:"
    echo ""
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
  fi

  echo ""
  echo "Initialize Agent Sail in a project:"
  echo ""
  echo "  cd /path/to/project"
  echo "  agentsail init"
  echo ""
  echo "Run the live load-test board:"
  echo ""
  echo "  agentsail loadtest tui --config agentsail.loadtest.yaml"
}

main "$@"
