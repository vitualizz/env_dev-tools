# EnvSetup

Your fast, personalized development environment installer.

## Philosophy

- Terminal-first workflow (Kitty + zsh + Neovim)
- No VSCode
- Cross-distro support (Arch, Debian, Fedora)
- AI-powered with opencode

## Quick Start

```bash
git clone https://github.com/vitualizz/envsetup.git
cd envsetup
docker compose run app
```

## Flow

1. **Language** - Select Spanish/English
2. **Theme** - Tokyo Night (your config), Nord, Dracula
3. **Tools** - Select what to install

## Commands

```bash
docker compose run app    # Run installer
docker compose run test   # Run tests
docker compose run shell  # Debug shell
```

## What's Included

| Category | Tools |
|----------|-------|
| Terminal | kitty (with your Tokyo Night config) |
| Shell | zsh, oh-my-zsh, starship, p10k |
| Editor | neovim |
| AI | opencode |
| Version Manager | mise (replaces rvm, volta, nvm) |
| Containers | docker, docker-compose |
| CLI Tools | fzf, bat, eza, bottom, yazi, zoxide, lazygit, direnv |
| Fonts | Hack Nerd Font |

## Configuration

- **Kitty**: `config/kitty/kitty.conf` + `color.ini`
- **Zsh**: `config/zsh/zshrc`, `zshenv`, `zprofile`
- **Tools**: `config/tools.yaml`

Install commands auto-detect your distro (apt, pacman, dnf, cargo).

## Development

```bash
go build -o envsetup ./cmd/envsetup/
go test ./...
go run ./cmd/envsetup/
```

## License

MIT