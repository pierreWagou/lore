---
title: Commands
weight: 1
---

## lore init

Create a `lore.toml` manifest interactively. Detects installed harnesses automatically.

```bash
lore init [flags]
```

| Flag | Description |
|---|---|
| `--mode` | Mode: `keeper` (commit skills) or `guest` (exclude from git) |
| `-g`, `--global` | Initialise global config at `~/.config/lore/lore.toml` |

---

## lore add

Add a skill to the manifest and install it.

```bash
lore add <source> [flags]
```

| Flag | Default | Description |
|---|---|---|
| `-g`, `--global` | false | Install globally |
| `--harness` | auto-detect | Comma-separated harness names (e.g. `opencode,claude`) |
| `-n`, `--name` | last path segment | Override the skill name |
| `-r`, `--ref` | `HEAD` | Git ref: branch, tag, or full SHA |
| `--all` | false | Install all discovered skills without prompting |

**Examples:**

```bash
# Specific skill path
lore add owner/repo/path/to/skill

# Scan repo, prompt for selection
lore add owner/repo

# Scan repo, install all
lore add owner/repo --all

# Pin to tag
lore add owner/repo/path --ref v1.2.0

# Target specific harness
lore add owner/repo/path --harness opencode

# Global install
lore add -g owner/repo/path
```

---

## lore create

Scaffold a new project skill in `.ai/skills/<name>/SKILL.md`.

```bash
lore create <name>
```

Creates `.ai/skills/<name>/SKILL.md` with a frontmatter template, adds a `[[dependencies]]` entry to `lore.toml`, and creates symlinks in configured harness dirs.

---

## lore import

Import skills from team harness directories into `.ai/skills/`. For use in guest mode after cloning a repo that has skills committed in a team harness dir.

```bash
lore import
```

Reads `team_harnesses` from `lore.toml` and copies any skills found there into `.ai/skills/`.

---

## lore export

Export skills from `.ai/skills/` to harness directories.

```bash
lore export [name] [flags]
```

| Flag | Description |
|---|---|
| `--harness` | Target harness (defaults to manifest harnesses) |
| `--all` | Export all skills |

---

## lore remove

Remove a skill from the manifest and uninstall from all harness directories.

```bash
lore remove <name> [-g]
```

| Flag | Description |
|---|---|
| `-g`, `--global` | Remove from global install |

---

## lore sync

Install all skills declared in `lore.toml`, using `lore.lock` for exact SHAs.

```bash
lore sync [-g] [--harness]
```

| Flag | Description |
|---|---|
| `-g`, `--global` | Sync global skills |
| `--harness` | Override target harnesses |

---

## lore list

Show installed skills with their source, ref, and locked commit.

```bash
lore list [-g]
```

| Flag | Description |
|---|---|
| `-g`, `--global` | List globally installed skills |

**Output:**

```
NAME                  SOURCE                                        REF    COMMIT
setup-second-brain    alan-eu/alan-skills/skills/setup-second-brain  main   a1b2c3d4
pdf                   anthropics/skills/pdf                         v2.1.0  f7e8d2c1
```

---

## lore harnesses

Detect installed harnesses on the current machine and show configured harnesses.

```bash
lore harnesses
```

### lore harnesses add

Add a harness to `lore.toml`.

```bash
lore harnesses add <name> [--team]
```

| Flag | Description |
|---|---|
| `--team` | Add as a team harness (`team_harnesses`) instead of personal |

### lore harnesses remove

Remove a harness from `lore.toml`.

```bash
lore harnesses remove <name>
```

---

## lore auth

Manage stored authentication tokens.

### lore auth add

```bash
lore auth add <host> <token>
```

**Examples:**

```bash
lore auth add github.com ghp_yourtoken
lore auth add gitlab.com glpat_yourtoken
lore auth add git.company.com yourpat
```

### lore auth list

```bash
lore auth list
```

**Output:**

```
HOST         TOKEN
github.com   ghp_...ef12
gitlab.com   glp_...ab34
```

### lore auth remove

```bash
lore auth remove <host>
```

---

## lore version

Print the lore version.

```bash
lore version
```

---

## lore completion

Generate shell completion scripts.

```bash
lore completion bash
lore completion zsh
lore completion fish
lore completion powershell
```
