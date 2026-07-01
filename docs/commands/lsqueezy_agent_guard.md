## lsqueezy agent guard

Emit host safety config that blocks destructive ops for an agent driving lsqueezy

### Synopsis

Classify every lsqueezy command (read / write / irreversible) using the same
annotations the MCP server uses, then emit safety config for the chosen agent host.

Reads are left free, writes require approval, and irreversible verbs (delete) are blocked.
Because it derives from the live tree, it stays correct as resources are added.

```
lsqueezy agent guard [flags]
```

### Examples

```
  lsqueezy agent guard --host claude-code
  lsqueezy agent guard --host codex
  lsqueezy agent guard --host opencode
```

### Options

```
      --all-writes    block all writes, not just irreversible ones
  -h, --help          help for guard
      --host string   agent host: claude-code|codex|opencode (default "claude-code")
      --write         write the config to the host's default path instead of stdout
```

### Options inherited from parent commands

```
      --account string    named account to use (env LEMONSQUEEZY_ACCOUNT)
      --all               fetch all pages (list commands)
      --base-url string   override the API base URL
      --columns strings   comma-separated columns to show
      --dry-run           print the equivalent curl and make no request
      --filter strings    client-side field=value filters (list commands)
      --jq string         gojq expression applied to the result before rendering
      --limit int         page size, 1-100 (list commands)
      --no-color          disable colored output
  -o, --output string     output format: table|json|yaml|csv
      --page int          page number, 1-based (list commands)
      --quiet             suppress non-essential chatter
      --show-token        reveal the API key in dry-run output
      --sort string       JSON:API sort field, prefix with - for desc (list commands)
  -v, --verbose           verbose request logging
```

### SEE ALSO

* [lsqueezy agent](lsqueezy_agent.md)	 - Generate AI-agent safety config from the live command tree

