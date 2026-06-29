## lsqueezy auth

Manage API credentials

### Options

```
  -h, --help   help for auth
```

### Options inherited from parent commands

```
      --all               fetch all pages (list commands)
      --base-url string   override the API base URL
      --columns strings   comma-separated columns to show
      --dry-run           print the equivalent curl and make no request
      --filter strings    client-side field=value filters (list commands)
      --limit int         page size, 1-100 (list commands)
      --no-color          disable colored output
  -o, --output string     output format: table|json|yaml|csv
      --page int          page number, 1-based (list commands)
      --profile string    config profile to use
      --quiet             suppress non-essential chatter
      --show-token        reveal the API key in dry-run output
      --sort string       JSON:API sort field, prefix with - for desc (list commands)
  -v, --verbose           verbose request logging
```

### SEE ALSO

* [lsqueezy](lsqueezy.md)	 - A polished CLI for the Lemon Squeezy e-commerce API
* [lsqueezy auth login](lsqueezy_auth_login.md)	 - Store an API key in the OS keyring and verify it
* [lsqueezy auth logout](lsqueezy_auth_logout.md)	 - Remove the stored API key for the active profile
* [lsqueezy auth status](lsqueezy_auth_status.md)	 - Show the active profile, base URL, and whether auth works

