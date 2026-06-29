# Auth & configuration

## Authentication

Lemon Squeezy uses a **bearer token** (an API key created in the dashboard under
Settings → API). Every request also sends `Accept`/`Content-Type: application/vnd.api+json`.

```bash
lsqueezy auth login --api-key <key>   # prompts if --api-key is omitted
lsqueezy auth status                   # alias: lsqueezy whoami
lsqueezy auth logout
```

`auth login` verifies the key against `GET /users/me` before storing it, so a bad key is
rejected immediately. The key is kept in the **OS keyring** (macOS Keychain / Linux Secret
Service / Windows Credential Manager) with an encrypted-file fallback on headless boxes —
never in the config file or repo.

### Test mode vs live

Lemon Squeezy distinguishes test and live data by the **kind of API key**, not the URL. Use a
separate profile per key:

```bash
lsqueezy --profile test auth login --api-key <test-key>
lsqueezy --profile test orders list
```

## Configuration precedence

`flag > env (LEMONSQUEEZY_*) > config file > default`.

| Setting    | Flag         | Env                     | Config file key |
|------------|--------------|-------------------------|-----------------|
| API key    | (keyring)    | `LEMONSQUEEZY_API_KEY`  | (never stored)  |
| Base URL   | `--base-url` | `LEMONSQUEEZY_BASE_URL` | `base_url`      |
| Output     | `-o`         | `LEMONSQUEEZY_OUTPUT`   | `output`        |
| Profile    | `--profile`  | `LEMONSQUEEZY_PROFILE`  | `active_profile`|

Config lives at `$XDG_CONFIG_HOME/lemon-squeezy-cli/config.yaml` (or
`~/.lemon-squeezy-cli/config.yaml`), written `0600` in a `0700` dir.

```bash
lsqueezy config path
lsqueezy config view                 # secrets redacted
lsqueezy config set output json
lsqueezy config use <profile>
lsqueezy config list-profiles
```

## First-run wizard & diagnostics

```bash
lsqueezy init       # capture key, write config, smoke-test against /users/me
lsqueezy doctor     # config / credentials / connectivity / auth / clock / version
lsqueezy doctor --json
```
