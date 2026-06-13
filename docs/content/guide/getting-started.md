---
title: Getting Started
weight: 2
---

## 1. Initialise a manifest

```bash
lore init
```

lore detects installed harnesses (opencode, claude) and creates a `lore.toml` in the current directory:

```toml
harnesses = ["opencode", "claude"]
```

Use `-g` for a global manifest at `~/.config/lore/lore.toml`:

```bash
lore init -g
```

---

## 2. Add a skill

```bash
# GitHub shorthand — specific path
lore add alan-eu/alan-skills/skills/setup-second-brain

# Pin to a tag
lore add alan-eu/alan-skills/skills/setup-second-brain --ref v1.0.0

# Install globally
lore add -g alan-eu/alan-skills/skills/setup-second-brain
```

Output:

```
installing setup-second-brain from alan-eu/alan-skills/skills/setup-second-brain...
  → opencode: ~/.config/opencode/skills/setup-second-brain
  → claude:   ~/.claude/skills/setup-second-brain
```

---

## 3. Scan a whole repo

If you omit the path, lore scans the repo for all skills:

```bash
lore add alan-eu/alan-skills
```

```
scanning https://github.com/alan-eu/alan-skills for skills...
found 4 skills:
  [1] setup-second-brain (alan-eu/alan-skills/skills/setup-second-brain)
  [2] pdf              (alan-eu/alan-skills/skills/pdf)
  [3] web-search       (alan-eu/alan-skills/skills/web-search)
  [4] code-review      (alan-eu/alan-skills/skills/code-review)
select skills to install (e.g. 1,3 or 'all'):
```

Use `--all` to skip the prompt:

```bash
lore add alan-eu/alan-skills --all
```

---

## 4. Sync from manifest

After cloning a repo or sharing a `lore.toml`, restore all skills:

```bash
lore sync
```

The `lore.lock` file pins exact commit SHAs — sync is reproducible.

---

## 5. List installed skills

```bash
lore list
```

```
NAME                  SOURCE                                        REF    COMMIT
setup-second-brain    alan-eu/alan-skills/skills/setup-second-brain  main   a1b2c3d4e5f6
```

---

## 6. Remove a skill

```bash
lore remove setup-second-brain
```

Removes from the manifest, lockfile, and all harness directories.
