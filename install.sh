#!/usr/bin/env bash
# Vitualizz DevStack — Bootstrap Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/vitualizz/vitualizz-devstack/main/install.sh | bash
set -euo pipefail

# --- Colors ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# --- Config ---
REPO="vitualizz/vitualizz-devstack"
INSTALL_DIR="${DEVSTACK_INSTALL_DIR:-/usr/local/bin}"
BIN_NAME="vitualizz-devstack"

info()    { echo -e "${CYAN}▸${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
warn()    { echo -e "${YELLOW}!${NC} $1"; }
error()   { echo -e "${RED}✗${NC} $1" >&2; }
fatal()   { error "$1"; exit 1; }

cleanup() { rm -f "$TMP_BIN" 2>/dev/null || true; }
trap cleanup EXIT

# --- Banner ---
echo -e "${BOLD}${CYAN}"
echo "  █▀▀ █░█ █▀▀ █░░ █▀▀ ▀█▀"
echo "  █░░ █▀█ █▀▀ █▄▄ █▄░ ░█░"
echo "  ▀▀▀ ▀░▀ ▀▀▀ ▀▀▀ ▀░▀ ░▀▀"
echo ""
echo "  ▀█▀ █▀█ █▀▀ █▀█"
echo "  ░█░ █▀▀ █▄▄ █▀▄"
echo "  ▀▀▀ ▀░░ ▀▀▀ ▀░▀"
echo ""
echo "  D e v S t a c k"
echo -e "${NC}"
echo

# --- Platform check ---
if [[ "$(uname -s)" != "Linux" ]]; then
  fatal "Vitualizz DevStack only supports Linux. macOS, Windows and BSD are not supported."
fi

# --- Architecture detection ---
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  GOARCH="x86_64" ;;
  aarch64) GOARCH="aarch64" ;;
  arm64)   GOARCH="aarch64" ;;
  *)       fatal "Unsupported architecture: $ARCH (only x86_64 and aarch64 supported)" ;;
esac

TMP_BIN=$(mktemp "/tmp/${BIN_NAME}.XXXXXX")

# --- Strategy: Download pre-built binary ---
download_binary() {
  info "Fetching latest release from ${REPO}..."

  # Use GitHub API to get latest release
  local latest_url
  latest_url=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | \
    grep -o '"browser_download_url": "[^"]*linux_'"${GOARCH}"'"' | \
    head -1 | \
    cut -d'"' -f4)

  if [[ -z "$latest_url" ]]; then
    error "No release binary found for linux/${ARCH}"
    warn ""
    warn "Falling back to Go build method..."
    return 1
  fi

  local version
  version=$(echo "$latest_url" | grep -oP 'v\K[0-9]+\.[0-9]+\.[0-9]+' || echo "latest")
  info "Downloading Vitualizz DevStack ${version} (linux/${ARCH})..."

  if ! curl -fsSL -o "$TMP_BIN" "$latest_url"; then
    error "Download failed"
    warn ""
    warn "Falling back to Go build method..."
    return 1
  fi

  chmod +x "$TMP_BIN"
  success "Binary downloaded"
  return 0
}

# --- Fallback: Build from source ---
build_from_source() {
  if ! command -v go &>/dev/null; then
    fatal "Go 1.24+ is required for source build. Install from https://go.dev/doc/install"
  fi

  local tmpdir
  tmpdir=$(mktemp -d "/tmp/devstack-build.XXXXXX")
  trap 'rm -rf "$tmpdir"' EXIT

  info "Cloning ${REPO}..."
  git clone --depth 1 "https://github.com/${REPO}.git" "$tmpdir" 2>/dev/null || \
    fatal "Failed to clone repository. Make sure git is installed."

  info "Building from source..."
  cd "$tmpdir"
  go build -o "$TMP_BIN" ./cmd/vitualizz-devstack/
  success "Binary built"
}

# --- Install ---
install_binary() {
  if [[ ! -w "$INSTALL_DIR" ]]; then
    warn "Cannot write to ${INSTALL_DIR}"
    warn "Please run with sudo: curl ... | sudo bash"
    fatal "Permission denied"
  fi

  cp "$TMP_BIN" "${INSTALL_DIR}/${BIN_NAME}"
  chmod +x "${INSTALL_DIR}/${BIN_NAME}"
  success "Installed to ${INSTALL_DIR}/${BIN_NAME}"
}

# --- Execute ---
if download_binary; then
  install_binary
else
  build_from_source
  install_binary
fi

echo
success "Vitualizz DevStack is ready!"
echo
echo "  Run: vitualizz-devstack"
echo "  CI:  vitualizz-devstack --ci"
echo
