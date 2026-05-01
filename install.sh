#!/usr/bin/env bash
# Vitualizz DevStack — Run-once Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/vitualizz/vitualizz-devstack/master/install.sh | bash
#
# Downloads the binary to /tmp, runs the TUI installer, then auto-deletes.
# No permanent binary is left on your system.
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
BIN_NAME="vitualizz-devstack"
LOG_DIR="${HOME}/.vitualizz-devstack"
LOG_FILE="${LOG_DIR}/install.log"

info()    { echo -e "${CYAN}▸${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
warn()    { echo -e "${YELLOW}!${NC} $1"; }
error()   { echo -e "${RED}✗${NC} $1" >&2; }
fatal()   { error "$1"; exit 1; }

# --- Logging ---
mkdir -p "$LOG_DIR"
echo "=== Install session started: $(date '+%Y-%m-%d %H:%M:%S') ===" >> "$LOG_FILE"

log_step() {
  local ts
  ts=$(date '+%H:%M:%S')
  echo "[$ts] $1" >> "$LOG_FILE"
}

log_error() {
  local ts
  ts=$(date '+%H:%M:%S')
  echo "[$ts] ERR: $1" >> "$LOG_FILE"
  if [[ -n "${2:-}" ]]; then
    echo "       Output: $2" >> "$LOG_FILE"
  fi
}

cleanup() { rm -f "$TMP_BIN" 2>/dev/null || true; }
trap cleanup EXIT

log_step "CMD: Platform check"

# --- Banner ---
echo "${BOLD}${CYAN}"
cat << 'BANNER'
           $$\   $$\                         $$\ $$\                     
           \__|  $$ |                        $$ |\__|                    
$$\    $$\ $$\ $$$$$$\   $$\   $$\  $$$$$$\  $$ |$$\ $$$$$$$$\ $$$$$$$$\ 
\$$\  $$  |$$ |\_$$  _|  $$ |  $$ | \____$$\ $$ |$$ |\____$$  |\____$$  |
 \$$\$$  / $$ |  $$ |    $$ |  $$ | $$$$$$$ |$$ |$$ |  $$$$ _/   $$$$ _/ 
  \$$$  /  $$ |  $$ |$$\ $$ |  $$ |$$  __$$ |$$ |$$ | $$  _/    $$  _/   
   \$  /   $$ |  \$$$$  |\$$$$$$  |\$$$$$$$ |$$ |$$ |$$$$$$$$\ $$$$$$$$\ 
    \_/    \__|   \____/  \______/  \_______|\__|\__|\________|\________|
                                                                         
  D e v S t a c k
BANNER
echo -e "${NC}"

# --- Platform check ---
log_step "CMD: uname -s ($(uname -s))"
if [[ "$(uname -s)" != "Linux" ]]; then
  log_error "Not Linux"
  fatal "Vitualizz DevStack only supports Linux. macOS, Windows and BSD are not supported."
fi

# --- Architecture detection ---
log_step "CMD: uname -m ($(uname -m))"
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
  log_step "CMD: Fetch releases from GitHub API"

  # Use GitHub API to get latest release
  local latest_url
  latest_url=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | \
    grep -o '"browser_download_url": "[^"]*linux_'"${GOARCH}"'"' | \
    head -1 | \
    cut -d'"' -f4)

  if [[ -z "$latest_url" ]]; then
    error "No release binary found for linux/${ARCH}"
    log_error "No binary URL found for linux/${ARCH}"
    warn ""
    warn "Falling back to Go build method..."
    return 1
  fi

  local version
  version=$(echo "$latest_url" | grep -oP 'v\K[0-9]+\.[0-9]+\.[0-9]+' || echo "latest")
  info "Downloading Vitualizz DevStack ${version} (linux/${ARCH})..."
  log_step "CMD: curl -o $TMP_BIN $latest_url"

  local output
  output=$(curl -fsSL -o "$TMP_BIN" "$latest_url" 2>&1) || {
    error "Download failed"
    log_error "Download failed" "$output"
    warn ""
    warn "Falling back to Go build method..."
    return 1
  }

  chmod +x "$TMP_BIN"
  success "Binary downloaded"
  log_step "OK: Binary downloaded ($(stat -c%s "$TMP_BIN" 2>/dev/null || echo "unknown") bytes)"
  return 0
}

# --- Fallback: Build from source ---
build_from_source() {
  log_step "CMD: Fallback to source build"
  if ! command -v go &>/dev/null; then
    log_error "Go not found"
    fatal "Go 1.24+ is required for source build. Install from https://go.dev/doc/install"
  fi

  local tmpdir
  tmpdir=$(mktemp -d "/tmp/devstack-build.XXXXXX")
  trap 'rm -rf "$tmpdir"' EXIT

  info "Cloning ${REPO}..."
  log_step "CMD: git clone --depth 1 ${REPO}"
  git clone --depth 1 "https://github.com/${REPO}.git" "$tmpdir" 2>/dev/null || \
    fatal "Failed to clone repository. Make sure git is installed."

  info "Building from source..."
  log_step "CMD: go build -o $TMP_BIN ./cmd/vitualizz-devstack/"
  cd "$tmpdir"
  go build -o "$TMP_BIN" ./cmd/vitualizz-devstack/
  success "Binary built"
  log_step "OK: Binary built from source"
}

# --- Execute ---
if download_binary; then
  :
else
  build_from_source
fi

echo
info "Starting Vitualizz DevStack..."
echo
info "📝 Log: ${LOG_FILE}"
echo

log_step "CMD: exec $TMP_BIN --self-destruct"

# Run the binary with --self-destruct flag so it deletes itself after use
exec "$TMP_BIN" --self-destruct
