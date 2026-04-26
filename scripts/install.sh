#!/usr/bin/env bash
# install.sh — Download and install the latest Agent Sail binary from GitHub Releases.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/LHANTAEK/cmux-260426-hackathon/mvp/scripts/install.sh | bash
set -euo pipefail

REPO="${AGENTSAIL_REPO:-LHANTAEK/cmux-260426-hackathon}"
INSTALL_DIR="${AGENTSAIL_INSTALL_DIR:-${HOME}/.local/bin}"
BINARY_NAME="agentsail"
TAG="${AGENTSAIL_VERSION:-latest}"
GITHUB_API="https://api.github.com/repos/${REPO}"
CODEX_HOME="${CODEX_HOME:-${HOME}/.codex}"
CURL_ARGS=()

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

warn() {
  echo "WARNING: $*" >&2
}

need() {
  command -v "$1" >/dev/null 2>&1 || error "'$1' is required but was not found"
}

append_curl_args() {
  CURL_ARGS=(-fsSL)
  if [[ -n "${GITHUB_TOKEN:-}" ]]; then
    CURL_ARGS+=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
  fi
  CURL_ARGS+=(-H "Accept: application/vnd.github+json")
}

detect_platform() {
  local os arch
  case "$(uname -s)" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    *) error "Unsupported OS: $(uname -s). Only macOS and Linux are supported." ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) error "Unsupported architecture: $(uname -m). Only amd64 and arm64 are supported." ;;
  esac
  echo "${os}-${arch}"
}

fetch_latest_tag() {
  if [[ "${TAG}" != "latest" ]]; then
    echo "${TAG}"
    return
  fi
  local tag
  append_curl_args
  tag=$(curl "${CURL_ARGS[@]}" "${GITHUB_API}/releases/latest" \
    | grep '"tag_name"' \
    | head -1 \
    | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
  [[ -n "${tag}" ]] || error "Could not fetch latest release tag for ${REPO}"
  echo "${tag}"
}

download() {
  local url="$1"
  local dest="$2"
  append_curl_args
  curl "${CURL_ARGS[@]}" --progress-bar -o "${dest}" "${url}" \
    || error "Download failed: ${url}"
}

try_download() {
  local url="$1"
  local dest="$2"
  append_curl_args
  curl "${CURL_ARGS[@]}" --progress-bar -o "${dest}" "${url}" >/dev/null 2>&1
}

verify_checksum() {
  local binary_file="$1"
  local checksum_file="$2"
  local binary_basename expected_line old_dir
  binary_basename="$(basename "${binary_file}")"
  expected_line="$(grep " ${binary_basename}$" "${checksum_file}" 2>/dev/null || true)"
  [[ -n "${expected_line}" ]] || error "No checksum entry found for ${binary_basename}"
  echo "${expected_line}" > "${TMP_DIR}/single.sha256"
  old_dir="$(pwd)"
  cd "${TMP_DIR}"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c single.sha256 --status \
      || error "SHA-256 checksum mismatch for ${binary_basename}"
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 -c single.sha256 --status \
      || error "SHA-256 checksum mismatch for ${binary_basename}"
  else
    error "sha256sum or shasum is required for checksum verification"
  fi
  cd "${old_dir}"
  info "Checksum verified"
}

install_codex_plugin() {
  if [[ "${AGENTSAIL_INSTALL_CODEX_PLUGIN:-1}" == "0" ]]; then
    return
  fi
  if ! command -v codex >/dev/null 2>&1; then
    return
  fi
  local tag="$1"
  info "Registering Agent Sail Codex plugin marketplace"
  if codex plugin marketplace add --ref "${tag}" "${REPO}" >/dev/null 2>&1; then
    return
  fi

  info "Refreshing existing Agent Sail Codex plugin marketplace"
  if codex plugin marketplace remove agentsail-marketplace >/dev/null 2>&1 \
    && codex plugin marketplace add --ref "${tag}" "${REPO}" >/dev/null 2>&1; then
    return
  fi

  warn "Codex plugin marketplace registration failed. Run manually:"
  warn "  codex plugin marketplace remove agentsail-marketplace"
  warn "  codex plugin marketplace add --ref ${tag} ${REPO}"
}

install_codex_skill() {
  if [[ "${AGENTSAIL_INSTALL_CODEX_SKILL:-1}" == "0" ]]; then
    return
  fi

  local tag="$1"
  local skill_dir="${CODEX_HOME}/skills/agentsail"
  local skill_file="${TMP_DIR}/agentsail-SKILL.md"
  local primary_url="https://raw.githubusercontent.com/${REPO}/${tag}/skills/agentsail/SKILL.md"
  local template_url="https://raw.githubusercontent.com/${REPO}/${tag}/internal/assets/templates/codex-plugin/skills/agentsail/SKILL.md"

  info "Installing Agent Sail Codex skill to ${skill_dir}"
  if ! try_download "${primary_url}" "${skill_file}"; then
    info "Repo skill not found; trying Codex plugin template"
    download "${template_url}" "${skill_file}"
  fi

  mkdir -p "${skill_dir}"
  cp "${skill_file}" "${skill_dir}/SKILL.md"
}

main() {
  need curl
  local platform tag asset_name base_url binary_path checksum_path
  platform="$(detect_platform)"
  tag="$(fetch_latest_tag)"
  asset_name="${BINARY_NAME}-${tag}-${platform}"
  base_url="https://github.com/${REPO}/releases/download/${tag}"

  TMP_DIR="$(mktemp -d)"
  binary_path="${TMP_DIR}/${asset_name}"
  checksum_path="${TMP_DIR}/checksums-sha256.txt"

  info "Installing Agent Sail ${tag} for ${platform}"
  download "${base_url}/${asset_name}" "${binary_path}"
  download "${base_url}/checksums-sha256.txt" "${checksum_path}"
  verify_checksum "${binary_path}" "${checksum_path}"

  mkdir -p "${INSTALL_DIR}"
  mv "${binary_path}" "${INSTALL_DIR}/${BINARY_NAME}"
  chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

  info "Installed ${INSTALL_DIR}/${BINARY_NAME}"
  "${INSTALL_DIR}/${BINARY_NAME}" --version
  install_codex_skill "${tag}"
  install_codex_plugin "${tag}"

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
