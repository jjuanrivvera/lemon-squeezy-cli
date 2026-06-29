## lsqueezy config

Inspect and edit configuration

### Options

```
  -h, --help   help for config
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
* [lsqueezy config list-profiles](lsqueezy_config_list-profiles.md)	 - List configured profiles
* [lsqueezy config path](lsqueezy_config_path.md)	 - Print the config file path
* [lsqueezy config set](lsqueezy_config_set.md)	 - Set a config value (base_url|output) for the active profile
* [lsqueezy config use](lsqueezy_config_use.md)	 - Switch the active profile
* [lsqueezy config view](lsqueezy_config_view.md)	 - Show the current config (secrets redacted)

