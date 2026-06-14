---
title: Installation
weight: 1
---

## Nix

```bash
nix profile install github:pierreWagou/nur#lore
```

---

## Homebrew

```bash
brew install pierreWagou/tap/lore
```

---

## Go

```bash
go install github.com/pierreWagou/lore@latest
```

Requires Go 1.26+. The binary is placed in `$GOPATH/bin` (typically `~/go/bin`).

---

## npm

```bash
npm install -g lore-agent
```

Downloads the platform-specific binary on install. Useful as a project dev dependency:

```bash
npm install --save-dev lore-agent
```

---

## pip

```bash
pip install lore-agent
```

Downloads the platform-specific binary on first run.

---

## Verify

```bash
lore version
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
