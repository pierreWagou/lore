---
title: Harnesses
weight: 4
---

A **harness** is an AI agent environment that lore can install skills into. Each harness has an adapter that knows the target directory structure and format.

---

## opencode

[opencode](https://opencode.ai) uses `SKILL.md` files in a `skills/` directory.

| Scope | Path |
|---|---|
| Global | `~/.config/opencode/skills/<name>/SKILL.md` |
| Project | `.opencode/skills/<name>/SKILL.md` |

**Detection:** lore detects opencode if `~/.config/opencode/` exists.

**Format:** pass-through — `SKILL.md` is installed as-is.

---

## claude

[Claude](https://claude.ai) (Anthropic) uses `SKILL.md` files in a `.claude/skills/` directory.

| Scope | Path |
|---|---|
| Global | `~/.claude/skills/<name>/SKILL.md` |
| Project | `.claude/skills/<name>/SKILL.md` |

**Detection:** lore detects claude if `~/.claude/` exists.

**Format:** pass-through — `SKILL.md` is installed as-is.

---

## Specifying harnesses

### In lore.toml

```toml
harnesses = ["opencode", "claude"]
```

### Via --harness flag

```bash
lore add owner/repo/path --harness opencode
lore sync --harness opencode,claude
```

### Auto-detection

When no harnesses are specified, lore detects installed harnesses and installs into all of them:

```bash
lore harnesses   # list detected harnesses
```

---

## Coming soon

| Harness | Format | Target path |
|---|---|---|
| cursor | `.mdc` with frontmatter | `.cursor/rules/<name>.mdc` |
| codex | `SKILL.md` | `~/.codex/skills/<name>/SKILL.md` |
| gemini | `SKILL.md` | `~/.gemini/skills/<name>/SKILL.md` |
| windsurf | `SKILL.md` | `~/.windsurf/skills/<name>/SKILL.md` |

Contributions welcome — see the `HarnessAdapter` interface in `internal/harness/harness.go`.
