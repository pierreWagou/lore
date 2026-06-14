---
title: lore.toml
weight: 3
---

`lore.toml` is lore's manifest file. It declares which skills to install and into which harnesses.

## Location

| Scope | Path |
|---|---|
| Project | `./lore.toml` (current directory) |
| Global | `~/.config/lore/lore.toml` |

Create with `lore init` or `lore init -g`.

---

## Schema

```toml
# keeper: .ai/skills/ committed, harness dirs gitignored
# guest:  .ai/skills/ ephemeral, personal harness dirs excluded via .git/info/exclude
mode = "keeper"

# Which harnesses to install skills into.
# Used by lore sync when no --harness flag is given.
# If omitted, lore auto-detects installed harnesses.
harnesses = ["opencode", "claude"]

# guest mode only: harness dirs committed by the team (read-only source for lore import)
team_harnesses = ["claude"]

# One [[dependencies]] block per skill.
[[dependencies]]
name   = "setup-second-brain"    # identifier used by lore remove / lore list
source = "alan-eu/alan-skills/skills/setup-second-brain"  # any source format
ref    = "main"                  # branch, tag, or full SHA

[[dependencies]]
name   = "pdf"
source = "anthropics/skills/pdf"
ref    = "v2.1.0"

[[dependencies]]
name   = "private-tool"
source = "git@github.com:my-org/private.git/tools/my-tool"
ref    = "abc123def456789"       # pin to exact commit
```

---

## Fields

### `mode`

Type: `string` — optional, default: `"keeper"`

Controls how lore manages file visibility in git.

| Value | Behaviour |
|---|---|
| `keeper` | `.ai/skills/` is committed; harness dirs gitignored via `.gitignore` |
| `guest` | `.ai/skills/` is ephemeral; personal harness dirs excluded via `.git/info/exclude` |

```toml
mode = "keeper"
```

---

### `team_harnesses`

Type: `[]string` — optional, guest mode only

Harness directories committed by the team. `lore import` scans these to populate `.ai/skills/`.

```toml
team_harnesses = ["claude"]
```

---

### `harnesses`

Type: `[]string` — optional

List of harness names to install skills into when running `lore sync`.

```toml
harnesses = ["opencode", "claude"]
```

If omitted or empty, `lore sync` auto-detects installed harnesses. Overridden by `--harness` flag.

Available harnesses: `opencode`, `claude` (more coming in future releases).

---

### `[[dependencies]]`

Repeatable block, one per skill.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Unique identifier for the skill |
| `source` | string | yes | Source handle — see [Source Formats]({{%/* ref "reference/source-formats" */%}}) |
| `ref` | string | no | Git ref: branch, tag, or full SHA (defaults to `HEAD`) |

{{% notice style="tip" %}}
Use a full SHA for production installs to guarantee reproducibility:

```toml
ref = "a1b2c3d4e5f6789012345678901234567890abcd"
```
{{% /notice %}}

---

## lore.lock

`lore.lock` is generated automatically by `lore add` and `lore sync`. **Do not edit it manually.**

```toml
# lore.lock — do not edit manually

[[entry]]
name         = "setup-second-brain"
source       = "alan-eu/alan-skills/skills/setup-second-brain"
commit       = "a1b2c3d4e5f6789012345678901234567890abcd"
content_hash = "sha256:deadbeef..."
resolved_at  = "2026-06-12T10:00:00Z"
```

Commit the lockfile to version-control so all contributors install identical skill versions.
