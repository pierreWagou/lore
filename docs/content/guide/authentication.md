---
title: Authentication
weight: 3
---

lore resolves credentials automatically. For most public repos, no configuration is needed.

## Resolution order (HTTPS)

lore tries these in sequence and uses the first match:

| Priority | Method | Scope |
|---|---|---|
| 1 | `LORE_<HOST>_TOKEN` env var | Any host |
| 2 | `GITHUB_TOKEN` env var | github.com only |
| 3 | `GH_TOKEN` env var | github.com only |
| 4 | `gh auth token` (gh CLI) | github.com only |
| 5 | `~/.config/gh/hosts.yml` | github.com only |
| 6 | Token stored via `lore auth add` | Any host |
| — | No auth (public repo) | Fallback |

The `LORE_<HOST>_TOKEN` format replaces `.` and `-` with `_` and uppercases:

| Host | Env var |
|---|---|
| `github.com` | `LORE_GITHUB_COM_TOKEN` |
| `gitlab.com` | `LORE_GITLAB_COM_TOKEN` |
| `git.company.com` | `LORE_GIT_COMPANY_COM_TOKEN` |

---

## GitHub (recommended setup)

If you already use the [gh CLI](https://cli.github.com/), lore uses its token automatically — no configuration needed:

```bash
gh auth login   # one-time setup
lore add github/private-repo/path/to/skill  # just works
```

Alternatively, set `GITHUB_TOKEN`:

```bash
export GITHUB_TOKEN=ghp_yourtoken
```

---

## Other hosts (GitLab, self-hosted)

Store a token with `lore auth add`:

```bash
lore auth add gitlab.com glpat_yourtoken
lore auth add git.company.com yourpersonalaccesstoken
```

Tokens are stored in `~/.config/lore/credentials.toml` with `0600` permissions.

```bash
# List stored tokens
lore auth list

# Remove a stored token
lore auth remove gitlab.com
```

{{% notice style="warning" %}}
Never commit `credentials.toml`. It is outside the project directory (`~/.config/lore/`) and should stay there.
{{% /notice %}}

---

## SSH

SSH is used automatically when the source handle starts with `git@`:

```bash
lore add git@github.com:org/private-repo.git/path/to/skill
```

lore tries in order:

1. **SSH agent** — uses `SSH_AUTH_SOCK` if set
2. **Key files** — auto-detects `~/.ssh/id_ed25519`, `id_rsa`, `id_ecdsa`, `id_dsa`

No configuration needed if your SSH agent is running or your key file is in the standard location.

---

## CI/CD

In GitHub Actions, `GITHUB_TOKEN` is available automatically:

```yaml
- name: Sync skills
  run: lore sync -g
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```
