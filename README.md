![header](https://capsule-render.vercel.app/api?type=waving&height=220&color=0:cba6f7,25:b4befe,50:89dceb,75:f5c2e7,100:f38ba8&text=lore&fontSize=80&fontColor=11111b&desc=chezmoi%20for%20agent%20skills&descSize=20&descAlignY=62&descAlign=50&fontAlignY=38&animation=fadeIn&fontAlign=50)

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)
[![Docs](https://img.shields.io/badge/docs-pierreWagou.github.io%2Flore-cba6f7)](https://pierreWagou.github.io/lore/)

</div>

---

**lore** is a personal agent skills package manager. It fetches skills from any git repository and installs them in each harness's native format — the way chezmoi manages dotfiles, lore manages your agent environment.

- Single Go binary, no runtime required
- Private repos via SSH, `gh` CLI, or stored tokens
- Unlimited path depth: `owner/repo/any/depth/path/to/skill`
- Lockfile pins exact commit SHAs for reproducibility
- Adapts to each harness's native format on install
- Global (`-g`) and project-scoped installs

## Architecture

```
 ┌───────────────────────────────────────────────┐
 │                   lore CLI                    │
 ├──────────┬──────────────┬──────────────────────┤
 │ resolver │  installer   │       harness        │  core
 ├──────────┼──────────────┼──────────┬───────────┤
 │  parse   │  fetch       │ opencode │  claude   │
 │  fetch   │  transform   │          │           │
 └──────────┴──────────────┴──────────┴───────────┘
       ↑                ↑
  lore.toml         lore.lock
  (manifest)        (pinned SHAs)
```

## Structure

```
lore/
├── cmd/lore/           entry point + CLI commands
├── internal/
│   ├── auth/           credential resolution (SSH, gh CLI, tokens)
│   ├── manifest/       lore.toml read/write
│   ├── lockfile/       lore.lock read/write
│   ├── resolver/       source handle parsing + git fetching
│   ├── scanner/        find SKILL.md files in a repo tree
│   ├── harness/        opencode + claude adapters
│   └── installer/      orchestrate fetch → transform → place → lock
└── wrappers/
    ├── npm/            thin npm wrapper (downloads binary on install)
    └── pip/            thin pip wrapper (downloads binary on first run)
```

## Getting Started

### Install

```bash
# Go
go install github.com/pierreWagou/lore@latest

# Homebrew (coming soon)
brew install pierreWagou/tap/lore

# npm
npm install -g lore-agent

# pip
pip install lore-agent
```

### Bootstrap a project

```bash
# Initialise a lore.toml (detects installed harnesses automatically)
lore init

# Initialise in guest mode (personal harness dirs excluded from git)
lore init --mode guest

# Add a skill
lore add owner/repo/path/to/skill

# Add a skill globally
lore add -g owner/repo/path/to/skill

# Sync all skills from lore.toml
lore sync
```

### Private repos

```bash
# GitHub: lore checks GITHUB_TOKEN, GH_TOKEN, and the gh CLI automatically.
# For other hosts, store a token:
lore auth add gitlab.com glpat_yourtoken
lore auth add git.company.com yourtoken
```

## Commands

| Command | Description |
|---|---|
| `lore init [--mode guest\|keeper] [-g]` | Create lore.toml (interactive) |
| `lore add <source> [-g] [--harness] [--ref] [--name] [--all]` | Add and install a skill |
| `lore create <name>` | Scaffold a new project skill in `.ai/skills/` |
| `lore import` | Import skills from team harness dirs into `.ai/skills/` |
| `lore export [name] [--harness] [--all]` | Export `.ai/skills/` to harness dirs |
| `lore remove <name> [-g]` | Remove from manifest and uninstall |
| `lore sync [-g] [--harness]` | Install all skills from lore.toml |
| `lore list [-g]` | List installed skills |
| `lore harnesses` | Detect installed harnesses |
| `lore harnesses add <name> [--team]` | Add a harness to lore.toml |
| `lore harnesses remove <name>` | Remove a harness from lore.toml |
| `lore auth add <host> <token>` | Store an auth token |
| `lore auth list` | List stored tokens |
| `lore auth remove <host>` | Remove a stored token |

## Source Formats

```bash
# GitHub shorthand — scans repo for all skills
lore add owner/repo

# GitHub shorthand — specific path
lore add owner/repo/path/to/skill

# Full HTTPS URL
lore add https://github.com/owner/repo/tree/main/path/to/skill

# SSH (private repo)
lore add git@github.com:owner/repo.git/path/to/skill

# Local path
lore add ./local-skills/my-skill
```

## lore.toml

```toml
mode     = "keeper"
harnesses = ["opencode", "claude"]

[[dependencies]]
name = "setup-second-brain"
source = "alan-eu/alan-skills/skills/setup-second-brain"
ref = "main"

[[dependencies]]
name = "pdf"
source = "anthropics/skills/pdf"
ref = "v2.1.0"
```

## Authentication

lore resolves credentials in this order for each fetch:

| Priority | Method |
|---|---|
| 1 | `LORE_GITHUB_COM_TOKEN` env var (host-specific override) |
| 2 | `GITHUB_TOKEN` / `GH_TOKEN` env var |
| 3 | `gh auth token` (gh CLI, if installed) |
| 4 | `~/.config/gh/hosts.yml` (gh CLI fallback) |
| 5 | Token stored via `lore auth add` |

SSH uses the system agent (`SSH_AUTH_SOCK`) and falls back to key files in `~/.ssh/`.

## Development

Dev tools (Go, goimports, lefthook, golangci-lint, hugo) are declared in `.mise.toml`.
Git hooks are installed automatically after `mise install`.

```bash
# Install all dev tools and activate git hooks
mise install

# Build
mise run build

# Test
mise run test

# Format (goimports)
mise run fmt

# Vet
mise run vet

# Lint
mise run lint

# Docs — local preview
mise run docs:dev

# Docs — production build
mise run docs:build
```

Install mise: https://mise.jdx.dev/getting-started.html

> **GitHub Pages setup (one-time):** go to Settings → Pages → Source: **GitHub Actions**

## Quick Reference

| Action | Command |
|---|---|
| Add skill globally | `lore add -g owner/repo/path` |
| Add to project | `lore add owner/repo/path` |
| Scan repo | `lore add owner/repo --all` |
| Pin to tag | `lore add owner/repo/path --ref v1.2.0` |
| Reinstall all | `lore sync` |
| Remove skill | `lore remove <name>` |
| Show installed | `lore list` |
| Store token | `lore auth add github.com <token>` |
| Build | `mise run build` |
| Test | `mise run test` |

## License

[MIT](LICENSE)
