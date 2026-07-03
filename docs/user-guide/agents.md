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

For `claude-code` the output includes two files: `.claude/settings.json` (permission rules
plus `PreToolUse` hook wiring) and `.claude/hooks/lsqueezy-guard.sh`. Pass `--write` to
install them under the project root (existing files are never overwritten). The hook exists
because permission rules are literal prefix patterns — on their own they would not catch a
path-invoked binary (`./bin/lsqueezy orders refund`), quote obfuscation
(`lsqueezy orders re""fund`), or a blocked command chained after `;`/`|`/`&&`. The hook
re-checks every Bash command with anchored command-position matching (an optional
`([^[:space:]]*/)?` path prefix covers `./bin/lsqueezy` and `/usr/local/bin/lsqueezy`) and
hard-blocks blocked MCP tools by exact name.

The `lsqueezy api` escape hatch requires approval as a whole, and its destructive HTTP
methods (`DELETE`/`PUT`/`POST`/`PATCH` — the method is the first positional argument) are
hard-blocked on the Bash surface. A `GET` whose path merely contains a word like `refund`
is not blocked.

### Known limitations (by design)

- **Variable indirection and shell aliases are not defeated** (`a=refund; lsqueezy orders
  $a 1`). For a hard guarantee run the agent MCP-only (no Bash tool) or in a read-only
  sandbox — the MCP branch of the hook is an exact tool-name match and cannot be obfuscated.
- **Conservative false positives deny, never allow**: a quoted full blocked command inside
  another program's arguments (e.g. `rg "lsqueezy orders refund" src/`) is denied because
  de-obfuscation strips quotes before matching. The failure direction is safe.
- Regenerate the config after upgrading `lsqueezy` so newly added commands are covered.
