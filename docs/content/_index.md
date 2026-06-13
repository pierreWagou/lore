---
title: lore
---

# lore

**Agent skills package manager** — install skills across opencode, claude, and more from any git source.

```bash
go install github.com/pierreWagou/lore@latest
lore add owner/repo/path/to/skill
```

---

{{% cards %}}
{{% card title="Getting Started" icon="rocket" link="guide/getting-started" %}}
Install lore and add your first skill in under a minute.
{{% /card %}}
{{% card title="Authentication" icon="lock" link="guide/authentication" %}}
Private repos via SSH, gh CLI, environment variables, or stored tokens.
{{% /card %}}
{{% card title="Command Reference" icon="book" link="reference/commands" %}}
Full documentation for every lore command and flag.
{{% /card %}}
{{% card title="Source Formats" icon="code" link="reference/source-formats" %}}
GitHub shorthand, HTTPS URLs, SSH, and local paths — all supported.
{{% /card %}}
{{% /cards %}}

---

## Why lore?

Most tools do "SKILL.md imperialism" — copying one format into every harness directory. lore adapts to each harness's native format on install, the way chezmoi manages dotfiles without imposing a canonical source format.

| | lore | microsoft/apm |
|---|---|---|
| Focus | Personal/global | Project/team |
| Runtime | Go binary | Python |
| Private repos | SSH + token + gh CLI | SSH + token |
| Lockfile | SHA + content hash | SHA |
| Harness-native output | Yes | Compile step |

lore and APM are complementary — lore is the tool you install globally to manage your personal agent environment.
