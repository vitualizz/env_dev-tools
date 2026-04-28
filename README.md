# EnvSetup

[![CI](https://github.com/vitualizz/envsetup/actions/workflows/ci.yml/badge.svg)](https://github.com/vitualizz/envsetup/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.24-blue?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Interactive TUI installer for a terminal-first development environment — Kitty, zsh, Neovim, and 40+ hand-picked tools, cross-distro (Arch, Debian, Alpine, Fedora).

## Quick Start

```bash
git clone https://github.com/vitualizz/envsetup.git
cd envsetup
docker compose run app
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

## How It Works

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
    all: cargo install ripgrep
```

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

## Development

```bash
# Run
go run ./cmd/envsetup/

# Test (all packages, race detector, coverage)
go test -race -cover ./...

# Build
go build -o envsetup ./cmd/envsetup/

# Docker (isolated environment)
docker compose run app     # run installer
docker compose run test    # run tests
docker compose run shell   # debug shell
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
