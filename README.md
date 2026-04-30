# Vitualizz DevStack

[![CI](https://github.com/vitualizz/vitualizz-devstack/actions/workflows/ci.yml/badge.svg)](https://github.com/vitualizz/vitualizz-devstack/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.24-blue?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Interactive TUI that installs and configures a complete terminal-first development environment on Linux — Kitty, zsh, Neovim, and 40+ hand-picked tools in one go.

> **Platform:** Linux only. macOS, Windows, and BSD are **not supported**. This is a Linux-native DevStack — it relies on distro package managers (pacman, apt, apk, dnf) and Linux-specific tooling.

## Quick Start

One command, that's it:

```bash
curl -fsSL https://raw.githubusercontent.com/vitualizz/vitualizz-devstack/main/install.sh | bash
```

The installer checks for Go 1.24+ first, then falls back to Docker. No manual setup needed.

### Manual Install

If you prefer to clone and run yourself:

```bash
git clone https://github.com/vitualizz/vitualizz-devstack.git
cd vitualizz-devstack
go run ./cmd/envsetup/              # with Go
docker compose run app              # or Docker (isolated)
```

## What Is This

A **DevStack** — not a dotfiles repo, not a collection of scripts. It's an opinionated, reproducible way to go from a bare Linux install to a fully configured developer environment.

```
Language select → Theme select → Tool select → Install
```

Each tool declares install commands per distro. The installer detects your distro at runtime and picks the right command, with a fallback chain: `exact distro → all → detection order → fallback`.

```yaml
# config/tools.yaml
- name: ripgrep
  install:
    arch: pacman -S ripgrep
    debian: apt-get install -y ripgrep
    alpine: apk add ripgrep
    all: cargo install ripgrep   # universal fallback
```

## Stack

| Category | Tools |
|----------|-------|
| **Terminal** | Kitty (+ Vitualizz color theme) |
| **Shell** | zsh, Oh My Zsh, Starship, Powerlevel10k, autosuggestions, syntax-highlighting, atuin |
| **Editor** | Neovim |
| **AI** | opencode |
| **Version managers** | mise, rustup, uv |
| **Containers** | Docker, docker-compose, lazydocker |
| **Git** | lazygit, delta, gh |
| **File tools** | eza, bat, yazi, fd, fzf, zoxide |
| **Search / replace** | ripgrep, sd |
| **Disk / process** | bottom, btop, duf, dust |
| **Docs** | tealdeer (tldr), glow, httpie |
| **Data** | jq, yq |
| **Info** | fastfetch, onefetch |
| **Fonts** | Hack, JetBrains Mono, Fira Code (Nerd Fonts) |
| **Theme** | Vitualizz |

## Supported Distros

- **Arch Linux** (pacman)
- **Debian / Ubuntu** (apt)
- **Alpine** (apk)
- **Fedora** (dnf)

Other distros may work if tools fall back to the `all` (cargo/universal) install path, but are not officially tested.

## Architecture

Hexagonal (Ports & Adapters) — UI and infrastructure never import each other, only through interfaces.

```
cmd/envsetup/
internal/
  domain/
    entities/     ← Tool, Theme, Distro, Category
    interfaces/   ← InstallerPort, ToolRepository (ports)
  usecases/       ← InstallTool, UninstallTool, BatchInstall, CheckStatus
  config/         ← YAML-based ToolRepository (adapter)
  infrastructure/
    executor/     ← ShellExecutor (runs commands, detects distro)
    installers/   ← ShellInstaller (wires executor to ports)
  ui/
    components/   ← Bubbletea app
    models/       ← AppModel (state machine)
config/
  tools.yaml      ← tool definitions
  kitty/          ← terminal config
i18n/             ← en/es translations
```

## Development Environment

### Prerequisites

- Go 1.24+
- Docker & Docker Compose (optional, for isolated testing)

### Run Locally

```bash
go run ./cmd/envsetup/
```

### Run Tests

```bash
# All packages with race detector and coverage
go test -race -cover ./...

# Verbose output
go test -v ./...
```

### Docker (Isolated Environment)

```bash
# Run the DevStack installer in a container
docker compose run app

# Run all tests in a clean container
docker compose run test

# Open a debug shell inside the build environment
docker compose run shell
```

The `app` service builds and runs the installer in an Alpine container — safe to experiment without touching your host system. The `test` service runs the full test suite in isolation.

### Build

```bash
go build -o envsetup ./cmd/envsetup/
```

## Adding a Tool

1. Add an entry to `config/tools.yaml`:

```yaml
- name: my-tool
  category: tools          # terminal | shell | editor | tools | container | fonts
  description: "Does X"
  install:
    arch: pacman -S my-tool
    debian: apt-get install -y my-tool
    all: cargo install my-tool   # universal fallback
  uninstall:
    all: rm ~/.local/bin/my-tool
  check: which my-tool
  enabled: true
  required: false
  depends_on:
    - rustup                # installed first if listed
```

2. Add translations to `i18n/locales/en.json` and `i18n/locales/es.json`.

## CI

Every push to `master` and every PR runs:

- `go build ./...`
- `go test -race -cover ./...`
- `golangci-lint`

## License

MIT

---

Con amor @vitualizz
