---
title: Installation
weight: 1
---

## Go (recommended)

```bash
go install github.com/pierreWagou/lore@latest
```

Requires Go 1.26+. The binary is placed in `$GOPATH/bin` (typically `~/go/bin`).

---

## Homebrew

```bash
brew install pierreWagou/tap/lore
```

{{% notice style="info" %}}
Homebrew tap coming soon. Use `go install` in the meantime.
{{% /notice %}}

---

## npm

```bash
npm install -g lore-agent
```

Downloads the platform-specific Go binary on install. Useful when you want lore as a project dev dependency:

```bash
npm install --save-dev lore-agent
```

---

## pip

```bash
pip install lore-agent
```

Downloads the platform-specific Go binary on first run.

---

## Verify

```bash
lore --version
```

---

## Shell completion

```bash
# bash
lore completion bash > /etc/bash_completion.d/lore

# zsh
lore completion zsh > "${fpath[1]}/_lore"

# fish
lore completion fish > ~/.config/fish/completions/lore.fish
```
