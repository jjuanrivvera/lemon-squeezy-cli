## lsqueezy mcp

MCP server management

### Synopsis

Manage MCP servers for AI assistants and code editors

### Options

```
  -h, --help[=true|false]   help for mcp
```

### Options inherited from parent commands

```
      --account string            named account to use (env LEMONSQUEEZY_ACCOUNT)
      --all[=true|false]          fetch all pages (list commands)
      --base-url string           override the API base URL
      --columns strings           comma-separated columns to show
      --dry-run[=true|false]      print the equivalent curl and make no request
      --filter strings            client-side field=value filters (list commands)
      --jq string                 gojq expression applied to the result before rendering
      --limit int                 page size, 1-100 (list commands)
      --no-color[=true|false]     disable colored output
  -o, --output string             output format: table|json|yaml|csv
      --page int                  page number, 1-based (list commands)
      --quiet[=true|false]        suppress non-essential chatter
      --show-token[=true|false]   reveal the API key in dry-run output
      --sort string               JSON:API sort field, prefix with - for desc (list commands)
  -v, --verbose[=true|false]      verbose request logging
```

### SEE ALSO

* [lsqueezy](lsqueezy.md)	 - A polished CLI for the Lemon Squeezy e-commerce API
* [lsqueezy mcp claude](lsqueezy_mcp_claude.md)	 - Manage Claude Desktop MCP servers
* [lsqueezy mcp cursor](lsqueezy_mcp_cursor.md)	 - Manage Cursor MCP servers
* [lsqueezy mcp start](lsqueezy_mcp_start.md)	 - Start the MCP server
* [lsqueezy mcp stream](lsqueezy_mcp_stream.md)	 - Stream the MCP server over HTTP
* [lsqueezy mcp tools](lsqueezy_mcp_tools.md)	 - Export tools as JSON
* [lsqueezy mcp vscode](lsqueezy_mcp_vscode.md)	 - Manage VSCode MCP servers

