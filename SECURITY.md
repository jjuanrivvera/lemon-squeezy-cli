# Security Policy

## Supported Versions

The latest released minor version receives security fixes.

| Version | Supported |
|---------|-----------|
| latest  | ✅        |

## Reporting a Vulnerability

Please report security issues privately via GitHub Security Advisories
(Security → Report a vulnerability) on this repository. Do not open a public issue
for security reports. We aim to acknowledge within 72 hours.

## Token Handling

- API keys are stored in the OS keyring (macOS Keychain / Linux Secret Service /
  Windows Credential Manager), never in the config file or repository.
- The `LEMONSQUEEZY_API_KEY` environment variable overrides stored credentials.
- `--dry-run` and `config view` redact the key by default; pass `--show-token` to reveal it.
- The MCP tool surface excludes `--api-key`, `--show-token`, `--profile`, and `--base-url`.
