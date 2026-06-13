---
title: Source Formats
weight: 2
---

lore accepts five source handle formats for `lore add` and in `lore.toml`.

---

## GitHub shorthand

```
owner/repo/path/to/skill
```

The first two segments are always `owner/repo`. Everything after is the path within the repo.

```bash
lore add alan-eu/alan-skills/skills/setup-second-brain
lore add anthropics/skills/pdf
```

Omit the path to scan the whole repo:

```bash
lore add alan-eu/alan-skills        # scans for all SKILL.md files
lore add alan-eu/alan-skills --all  # installs all without prompting
```

{{% notice style="info" %}}
Shorthand always resolves against `github.com`. For other hosts, use a full HTTPS URL.
{{% /notice %}}

---

## Full HTTPS URL

```
https://host/owner/repo/tree/<ref>/path/to/skill
```

```bash
lore add https://github.com/owner/repo/tree/main/path/to/skill
lore add https://gitlab.com/owner/repo/tree/v1.0.0/path/to/skill
lore add https://git.company.com/org/repo                        # scan
```

The `tree/<ref>/` segment is parsed as the git ref. An explicit `--ref` flag overrides it:

```bash
lore add https://github.com/owner/repo/tree/main/path --ref v2.0.0
# uses v2.0.0, not main
```

---

## SSH (private repos)

```
git@host:owner/repo.git[/path/to/skill]
```

The `.git` suffix is required. Everything after `.git/` is the subpath.

```bash
lore add git@github.com:my-org/private-skills.git/path/to/skill
lore add git@github.com:my-org/private-skills.git   # scan whole repo
```

lore uses your SSH agent or `~/.ssh` key files automatically. See [Authentication]({{%/* ref "guide/authentication" */%}}) for details.

---

## Local path

```
./relative/path/to/skill
/absolute/path/to/skill
```

```bash
lore add ./local-skills/my-skill
lore add /home/user/shared-skills/tool
```

Local skills are installed immediately without git. No `lore.lock` entry is created (no commit SHA to record).

---

## In lore.toml

All formats work as the `source` field in `lore.toml`:

```toml
[[dependencies]]
name   = "second-brain"
source = "alan-eu/alan-skills/skills/setup-second-brain"
ref    = "main"

[[dependencies]]
name   = "private-tool"
source = "git@github.com:my-org/skills.git/tools/my-tool"
ref    = "abc123def456"

[[dependencies]]
name   = "local-dev"
source = "./local-skills/my-skill"
ref    = ""
```
