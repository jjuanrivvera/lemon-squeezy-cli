# Installation

## Homebrew

```bash
brew install jjuanrivvera/lemon-squeezy-cli/lemon-squeezy-cli
```

## Go

```bash
go install github.com/jjuanrivvera/lemon-squeezy-cli@latest
```

The binary is named `lsqueezy` (not `ls`, which shadows coreutils, nor `lsq`, which clashes
with the Logseq CLI).

## Release binaries

Prebuilt archives for macOS, Linux, and Windows are attached to each
[GitHub release](https://github.com/jjuanrivvera/lemon-squeezy-cli/releases). Download the one
for your platform, extract it, and put `lsqueezy` on your `PATH`.

## Shell completion

```bash
# bash
lsqueezy completion bash > /usr/local/etc/bash_completion.d/lsqueezy
# zsh
lsqueezy completion zsh > "${fpath[1]}/_lsqueezy"
# fish
lsqueezy completion fish > ~/.config/fish/completions/lsqueezy.fish
```

Completions are also shipped inside the release archives and the Homebrew formula.

## Verify

```bash
lsqueezy version
lsqueezy doctor
```

`doctor` checks that the config is readable, the key resolves, connectivity works, and the
clock is sane — run it first whenever something misbehaves.
