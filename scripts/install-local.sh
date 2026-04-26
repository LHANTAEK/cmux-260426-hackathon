#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR="${AGENTSAIL_INSTALL_DIR:-${HOME}/.local/bin}"
BINARY_NAME="agentsail"
CODEX_HOME="${CODEX_HOME:-${HOME}/.codex}"

info() {
  printf '==> %s\n' "$*"
}

warn() {
  printf 'WARNING: %s\n' "$*" >&2
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
if [[ "${AGENTSAIL_INSTALL_CODEX_SKILL:-1}" != "0" ]]; then
  CODEX_SKILL_SRC="${ROOT_DIR}/skills/agentsail"
  if [[ ! -f "${CODEX_SKILL_SRC}/SKILL.md" ]]; then
    CODEX_SKILL_SRC="${ROOT_DIR}/internal/assets/templates/codex-plugin/skills/agentsail"
  fi
  if [[ -f "${CODEX_SKILL_SRC}/SKILL.md" ]]; then
    CODEX_SKILL_DIR="${CODEX_HOME}/skills/agentsail"
    info "Installing Agent Sail Codex skill to ${CODEX_SKILL_DIR}"
    mkdir -p "${CODEX_SKILL_DIR}"
    cp "${CODEX_SKILL_SRC}/SKILL.md" "${CODEX_SKILL_DIR}/SKILL.md"
  else
    warn "Agent Sail Codex skill source not found; skipping skill install"
  fi
fi
if [[ "${AGENTSAIL_INSTALL_CODEX_PLUGIN:-1}" != "0" ]] && command -v codex >/dev/null 2>&1; then
  info "Registering local Agent Sail Codex plugin marketplace"
  if ! codex plugin marketplace add "${ROOT_DIR}" >/dev/null 2>&1; then
    info "Refreshing existing local Agent Sail Codex plugin marketplace"
    codex plugin marketplace remove agentsail-marketplace >/dev/null 2>&1 \
      && codex plugin marketplace add "${ROOT_DIR}" >/dev/null 2>&1 \
      || warn "Codex plugin marketplace registration failed. Run: codex plugin marketplace remove agentsail-marketplace && codex plugin marketplace add ${ROOT_DIR}"
  fi
fi
if [[ ":${PATH}:" != *":${INSTALL_DIR}:"* ]]; then
  echo "WARNING: ${INSTALL_DIR} is not in PATH."
  echo "Add this to your shell profile:"
  echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi

echo ""
echo "Initialize Agent Sail in a project:"
echo "  cd /path/to/project"
echo "  agentsail init"
