# Vitualizz DevStack

[![CI](https://github.com/vitualizz/vitualizz-devstack/actions/workflows/ci.yml/badge.svg)](https://github.com/vitualizz/vitualizz-devstack/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.24-blue?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Interactive TUI that installs and configures a complete terminal-first development environment on Linux — Kitty, zsh, Neovim, and 40+ hand-picked tools in one go.

> **Platform:** Linux only. macOS, Windows, and BSD are **not supported**. This is a Linux-native DevStack — it relies on distro package managers (pacman, apt, apk, dnf) and Linux-specific tooling.

## Quick Start

One command, no dependencies needed:

```bash
curl -fsSL https://raw.githubusercontent.com/vitualizz/vitualizz-devstack/master/install.sh | sudo bash
```

The installer downloads a pre-compiled binary from the latest GitHub release. No Go, no Docker, no compilation — just runs.

### Manual Install

If you prefer to build from source:

```bash
git clone https://github.com/vitualizz/vitualizz-devstack.git
cd vitualizz-devstack
go run ./cmd/vitualizz-devstack/     # interactive TUI
go run ./cmd/vitualizz-devstack/ --ci  # headless mode
```

Or use Docker:

```bash
docker compose run app               # CI mode (headless)
docker compose run app ./vitualizz-devstack --tui  # interactive
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
cmd/vitualizz-devstack/
    config/          ← embedded config files (tools.yaml, kitty/, zsh/)
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
i18n/             ← en/es translations
```

> **Note**: `config/` (tools.yaml, kitty/, zsh/) lives inside `cmd/vitualizz-devstack/config/` and is embedded into the binary at build time via `go:embed`.
```

## Development Environment

### Prerequisites

- Go 1.24+
- Docker & Docker Compose (optional, for isolated testing)
- Vagrant + VirtualBox/libvirt (optional, for multi-distro testing)

### Run Locally

```bash
go run ./cmd/vitualizz-devstack/
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
# Run the DevStack installer in CI mode (headless)
docker compose run app

# Run in interactive TUI mode
docker compose run app ./vitualizz-devstack --tui

# Run all tests in a clean container
docker compose run test

# Open a debug shell inside the build environment
docker compose run shell
```

The `app` service builds and runs the installer in CI mode by default — no TUI, clean output, auto-detects Docker and skips incompatible tools (kitty, docker, etc.).

### Vagrant (Multi-Distro Testing)

Boot real Linux VMs to test the installer across distros:

```bash
vagrant up ubuntu    # Ubuntu 22.04 LTS (~2 min)
vagrant up arch      # Arch Linux (~3 min)

vagrant ssh ubuntu   # SSH into the VM
# Inside the VM:
cd /vagrant
go run ./cmd/vitualizz-devstack/

vagrant destroy -f   # Clean up when done
```

Both VMs come with 2 vCPUs and 2GB RAM. The project root is synced to `/vagrant` so you can build and run the installer directly.

| Distro | Box | Provider |
|--------|-----|----------|
| Ubuntu 22.04 | `ubuntu/jammy64` | VirtualBox / libvirt |
| Arch Linux | `archlinux/archlinux` | VirtualBox / libvirt |

### Build

```bash
go build -o vitualizz-devstack ./cmd/vitualizz-devstack/

# Run in CI mode (headless, no TUI)
./vitualizz-devstack --ci

# Run in TUI mode (interactive)
./vitualizz-devstack
```

### CI Mode

The `--ci` flag runs the installer without a TUI — designed for:
- **Docker testing** — no TTY required, clean output
- **CI/CD pipelines** — exit code 1 if any tool fails
- **Quick verification** — see what installs and what doesn't

## Releases

Binary releases are built automatically via GitHub Actions + GoReleaser when a tag is pushed:

```bash
git tag v1.0.0 && git push origin v1.0.0
```

This triggers the `release.yml` workflow which:
1. Builds static binaries for `linux/amd64` and `linux/arm64`
2. Creates a GitHub Release with changelog
3. Uploads checksums

No Go installation needed — `install.sh` downloads the pre-built binary directly.

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
