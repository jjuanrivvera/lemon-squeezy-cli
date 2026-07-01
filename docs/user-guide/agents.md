# AI agents (MCP server + guard)

`lsqueezy` can drive itself from an AI agent two ways: as an MCP server, and by generating
agent-safety config for host tools.

## MCP server

```bash
lsqueezy mcp start
```

This runs `lsqueezy` as a [Model Context Protocol](https://modelcontextprotocol.io) server.
It walks the command tree and exposes each runnable command as an annotated tool
(`lsqueezy_<resource>_<verb>`), replaying the real cobra command on invocation — so every tool
reuses the same client, keyring, account, and `--dry-run`.

Setup and secret commands are **excluded** from the MCP surface (`auth`, `config`, `init`,
`alias`, `doctor`, `agent`), and the account/secret flags (`--account`, the hidden `--profile`
alias, `--api-key`, `--show-token`, `--base-url`) are dropped from every tool schema. The server
stays on whatever account was active at startup, so an agent can't switch accounts or read the
key.

Install the server into a host's config:

```bash
lsqueezy mcp claude enable
lsqueezy mcp cursor enable
lsqueezy mcp vscode enable
```

## Agent guard

`agent guard` generates host-specific safety rules from the live command tree, classifying
every command read-only / write / destructive:

```bash
lsqueezy agent guard --host claude-code   # refund/cancel/delete/deactivate denied; writes ask
lsqueezy agent guard --host codex
lsqueezy agent guard --host opencode
```

Read verbs (`list`, `get`, `current-usage`) are allowed, writes (`create`, `update`, `archive`)
require approval, and irreversible verbs (`refund`, `cancel`, `delete`, `deactivate`) are hard
blocked. Custom verbs are treated as destructive unless a resource marks them read-only, so a
new command is safe by default.
