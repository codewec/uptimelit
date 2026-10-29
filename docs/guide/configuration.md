# Configuration

Monitor and notifier settings live in SQLite and are managed through the UI or [REST API](api.md). Runtime configuration comes from process environment variables. `-addr` and `-db` override the corresponding environment values.

## Environment variables

| Variable | Default | Description |
| --- | --- | --- |
| `UPTIMELIT_ADDR` | `:8080` | HTTP listen address; CLI override: `-addr` |
| `UPTIMELIT_DB` | `./uptimelit.db` | SQLite database path; CLI override: `-db` |
| `JWT_SECRET` | Generated and persisted in SQLite | Optional session signing key, at least 32 bytes |
| `OIDC_ISSUER` | Empty | Provider issuer URL |
| `OIDC_CLIENT_ID` | Empty | OIDC client ID |
| `OIDC_CLIENT_SECRET` | Empty | OIDC client secret |
| `OIDC_REDIRECT_URL` | Empty | Absolute HTTP(S) callback URL, required when OIDC is enabled |
| `OIDC_BUTTON_TITLE` | `Sign in with OIDC` | Provider login button text |
| `DISABLE_PASSWORD_LOGIN` | `false` | Disable password login when OIDC is configured |
| `OIDC_INSECURE_SKIP_TLS_VERIFY` | `false` | Bypass TLS certificate validation for OIDC requests only |

Boolean switches are enabled by the literal value `true`. The Docker image sets `UPTIMELIT_DB=/data/uptimelit.db`; the included Compose file also specifies this path.

Issuer and client ID must be supplied together. Partial OIDC configuration, an invalid callback URL or an explicitly supplied signing secret shorter than 32 bytes prevents startup with a configuration error.

## Environment files

The Go binary does not load `.env` files automatically. For a native process, export variables in its environment. For `docker run`, use `-e` or `--env-file`.

The [Compose file](../../compose.yml) explicitly forwards its listed authentication variables from the shell or project `.env`. Add or edit a service environment entry if you want to override other runtime values such as `UPTIMELIT_ADDR` or `UPTIMELIT_DB`.

An optional [.env.example](../../.env.example) is included:

```bash
cp .env.example .env
# Edit .env, then recreate the service.
docker compose up -d
```

## Signing key

Without `JWT_SECRET`, UptimeLit generates a key and stores it in the database. Persisting that database preserves the key across restarts. When setting a key explicitly, use at least 32 bytes, for example a value generated with `openssl rand -hex 32`.

Changing the effective key invalidates existing tokens. Preserve both the database and any explicit deployment key when making [backups](backups.md).

## Reverse proxy

Forward the public `Host` and `X-Forwarded-Proto` headers. Use the public HTTPS URL for `OIDC_REDIRECT_URL` and configure the exact same callback in the provider. Requests with session cookies must keep the UI and API on the same origin for CSRF protection.

The server marks cookies secure when HTTPS is detected from TLS, the forwarded protocol or an HTTPS OIDC callback. See [OIDC setup](oidc.md) for provider configuration and TLS options.
