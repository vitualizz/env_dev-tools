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
BRANCH="main"
RAW_URL="https://raw.githubusercontent.com/${REPO}/${BRANCH}"

info()    { echo -e "${CYAN}▸${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
warn()    { echo -e "${YELLOW}!${NC} $1"; }
error()   { echo -e "${RED}✗${NC} $1" >&2; }
fatal()   { error "$1"; exit 1; }

# --- Banner ---
echo -e "${BOLD}${CYAN}"
echo "  __     ___     _            ____                  "
echo "  \ \   / (_)___| |_ ___  _ __/ ___|  ___ __ _ _ __ "
echo "   \ \ / /| / __| __/ _ \| '__\___ \ / __/ _\` | '__|"
echo "    \ V / | \__ \ || (_) | |   ___) | (_| (_| | |   "
echo "     \_/  |_|___/\__\___/|_|  |____/ \___\__,_|_|   "
echo -e "${NC}"
echo -e "${BOLD}Vitualizz DevStack${NC} — Terminal-first dev environment"
echo

# --- Platform check ---
if [[ "$(uname -s)" != "Linux" ]]; then
  fatal "Vitualizz DevStack only supports Linux. macOS, Windows and BSD are not supported."
fi

# --- Temp directory ---
TMPDIR=$(mktemp -d "/tmp/devstack.XXXXXX")
trap 'rm -rf "$TMPDIR"' EXIT

clone_repo() {
  info "Cloning ${REPO}@${BRANCH}..."
  git clone --depth 1 --branch "$BRANCH" \
    "https://github.com/${REPO}.git" "$TMPDIR" 2>/dev/null || \
    fatal "Failed to clone repository. Make sure git is installed."
  success "Repository cloned."
}

# --- Strategy 1: Go (preferred) ---
try_go() {
  if ! command -v go &>/dev/null; then
    return 1
  fi

  GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
  GO_MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
  GO_MINOR=$(echo "$GO_VERSION" | cut -d. -f2)

  if [[ "$GO_MAJOR" -lt 1 || ("$GO_MAJOR" -eq 1 && "$GO_MINOR" -lt 24) ]]; then
    warn "Go $GO_VERSION found, but 1.24+ is required."
    return 1
  fi

  clone_repo
  echo
  info "Building DevStack with Go $GO_VERSION..."
  cd "$TMPDIR"
  go run ./cmd/envsetup/
}

# --- Strategy 2: Docker ---
try_docker() {
  if ! command -v docker &>/dev/null; then
    return 1
  fi

  clone_repo
  echo
  info "Running DevStack in Docker (CI mode)..."
  cd "$TMPDIR"

  if command -v docker-compose &>/dev/null; then
    docker-compose run --rm app
  elif docker compose version &>/dev/null 2>&1; then
    docker compose run --rm app
  else
    fatal "Docker is installed but docker-compose is not available."
  fi
}

# --- Execution ---
info "Checking prerequisites..."

if try_go; then
  echo
  success "DevStack launched with Go."
  exit 0
fi

if try_docker; then
  echo
  success "DevStack launched with Docker."
  exit 0
fi

fatal "No suitable runtime found.

Vitualizz DevStack requires one of:
  • Go 1.24+   — https://go.dev/doc/install
  • Docker     — https://docs.docker.com/engine/install/

Install one and try again."
