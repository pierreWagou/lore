---
title: Commands
weight: 1
---

## lore init

Create a `lore.toml` manifest interactively. Detects installed harnesses automatically.

```bash
lore init [-g]
```

| Flag | Description |
|---|---|
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
| `-t`, `--target` | auto-detect | Comma-separated harness names (e.g. `opencode,claude`) |
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
lore add owner/repo/path --target opencode

# Global install
lore add -g owner/repo/path
```

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
lore sync [-g] [-t target]
```

| Flag | Description |
|---|---|
| `-g`, `--global` | Sync global skills |
| `-t`, `--target` | Override target harnesses |

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

## lore targets

Detect installed harnesses on the current machine.

```bash
lore targets
```

**Output:**

```
detected harnesses:
  opencode
  claude
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
