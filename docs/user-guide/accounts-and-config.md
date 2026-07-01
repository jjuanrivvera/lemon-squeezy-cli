# Accounts & configuration

## Precedence

Settings resolve in this order, highest first:

```text
flag  >  env (LEMONSQUEEZY_*)  >  config file  >  default
```

## Accounts (named profiles)

A Lemon Squeezy profile is an account — one API key. One machine can hold several (for example
a live key and a test-mode key) and switch between them. The selector flag is `--account`:

```bash
lsqueezy --account test orders list
lsqueezy --account live subscriptions list
```

`--profile` still works as a hidden alias so older scripts don't break, and
`LEMONSQUEEZY_ACCOUNT` (or the legacy `LEMONSQUEEZY_PROFILE`) sets it from the environment.

Each account stores its own key in the OS keyring, so switching accounts switches credentials.

```bash
lsqueezy config use test         # create/switch the active account
lsqueezy --account test auth login --api-key test_xxx
lsqueezy config list-profiles    # * marks the active one
```

## Config file

Non-secret settings (base URL, default output, active account) live at:

```text
$XDG_CONFIG_HOME/lemon-squeezy-cli/config.yaml
# or ~/.lemon-squeezy-cli/config.yaml
```

The API key is **never** stored here — only in the keyring.

```bash
lsqueezy config path
lsqueezy config view                                   # secrets redacted
lsqueezy config set output json
lsqueezy config set base_url https://api.lemonsqueezy.com/v1
```

## Environment overrides

| Variable                 | Overrides                                  |
|--------------------------|--------------------------------------------|
| `LEMONSQUEEZY_API_KEY`   | the stored key (wins over the keyring)     |
| `LEMONSQUEEZY_ACCOUNT`   | the active account (`--account`)           |
| `LEMONSQUEEZY_BASE_URL`  | the API base URL (`--base-url`)            |
| `LEMONSQUEEZY_OUTPUT`    | the default output format (`-o`)           |
| `NO_COLOR`               | disables table color                       |
