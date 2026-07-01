# Authentication

Lemon Squeezy uses a bearer API key. Create one under **Settings → API** in your Lemon Squeezy
dashboard, then hand it to `lsqueezy`.

## Store the key in the OS keyring

```bash
lsqueezy auth login --api-key eyJ0eXAiOi...
```

The key is verified against `GET /users/me` before it is saved, so a bad key never lands in
the keyring. On success you'll see the authenticated user. Omit `--api-key` to be prompted
instead of passing it on the command line (which would leak into your shell history).

The key lives in your OS keyring (macOS Keychain, the Secret Service on Linux, or the Windows
Credential Manager), with an encrypted-file fallback for headless boxes. It is **never** written
to the config file.

## Headless / CI

Set the key in the environment instead of the keyring:

```bash
export LEMONSQUEEZY_API_KEY=eyJ0eXAiOi...
lsqueezy stores list
```

`LEMONSQUEEZY_API_KEY` takes precedence over the keyring, so it's the right lever for CI jobs
and containers.

## Check and remove

```bash
lsqueezy auth status      # or: lsqueezy whoami
lsqueezy auth logout      # remove the stored key for the active account
```

## The License API is different

The License API (`lsqueezy license activate|validate|deactivate`) works with just a license
key — the store API key is optional, since license checks run on customer machines. See
[Creating & updating](../user-guide/writing-data.md) and the command reference for details.
