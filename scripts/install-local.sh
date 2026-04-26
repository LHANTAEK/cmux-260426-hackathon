#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR="${AGENTSAIL_INSTALL_DIR:-${HOME}/.local/bin}"
BINARY_NAME="agentsail"

info() {
  printf '==> %s\n' "$*"
}

case "$(uname -s)" in
  Darwin|Linux) ;;
  *) echo "Unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac

mkdir -p "${INSTALL_DIR}"
info "Building ${BINARY_NAME}"
GOCACHE="${ROOT_DIR}/.gocache" GOMODCACHE="${ROOT_DIR}/.gomodcache" \
  go build -o "${INSTALL_DIR}/${BINARY_NAME}" "${ROOT_DIR}/cmd/agentsail"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

info "Installed ${INSTALL_DIR}/${BINARY_NAME}"
if [[ ":${PATH}:" != *":${INSTALL_DIR}:"* ]]; then
  echo "WARNING: ${INSTALL_DIR} is not in PATH."
  echo "Add this to your shell profile:"
  echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi

echo ""
echo "Initialize Agent Sail in a project:"
echo "  cd /path/to/project"
echo "  agentsail init"

