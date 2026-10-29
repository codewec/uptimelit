<h1><img src="uptimelit.png" width="36" height="36" align="center" style="vertical-align: middle; margin-right: 8px;"> UptimeLit</h1>

A self-hosted uptime monitor for HTTP/HTTPS and TCP services, written in Go with a SQLite database and an embedded web interface. Each user has a private workspace for monitors and notification channels.

[![Go Version](https://img.shields.io/badge/go-1.26.2+-00ADD8.svg)](go.mod)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

UptimeLit was originally based on [TinyPulse](https://github.com/AkaCoder404/tinypulse) by AkaCoder404.

> The project is in beta.

## Features

- HTTP/HTTPS and TCP monitoring with configurable check intervals and failure thresholds.
- Current service status, response times, recent check history and 24-hour uptime statistics.
- SQLite storage with WAL mode, automatic schema migrations and 90-day check retention.
- Telegram, Pushover and Redis notifications for outages and recoveries.
- Private monitors, history and notification channels for every user, including administrators.
- Administrator-managed accounts with `admin` and `user` roles; no public registration.
- Password and OpenID Connect (OIDC) login, revocable sessions and profile management.
- A REST API for managing monitors, notifiers and accounts.
- Docker deployment or a standalone Go binary; no separate database server or Node.js runtime required.

## Quick start

### Docker Compose

Clone the repository and start the included [Compose service](compose.yml):

```bash
git clone https://github.com/codewec/uptimelit.git
cd uptimelit
docker compose up -d --build
```

Open [http://localhost:8080](http://localhost:8080). On an empty database, sign in with a new username and password to create the first administrator. Passwords must contain 8–72 bytes; usernames accept 1–64 letters, digits, dots, underscores or dashes. If OIDC is configured, the first successful OIDC sign-in can create the administrator instead.

After the first account exists, only an administrator can create additional users through **Users**. Unknown accounts cannot sign in or register.

The container runs as UID/GID `10001:10001`. The `uptimelit-data` named volume stores the database and its WAL files in `/data`. Recreating the container preserves these files. `docker compose down` keeps the volume; `docker compose down -v` deletes it.

To update a deployment built from source:

```bash
git pull
docker compose up -d --build
```

See the [deployment guide](docs/deployment_and_updates.md) for deployment and database backup notes.

### Docker image from GHCR

The [GitHub Actions workflow](.github/workflows/docker-dev.yml) builds the Dockerfile and publishes `ghcr.io/codewec/uptimelit:dev` on pushes to `main`. It also supports manual runs for `main`. The `dev` tag floats to the latest published build.

To run the published image:

```bash
docker pull ghcr.io/codewec/uptimelit:dev
docker run -d --name uptimelit --restart unless-stopped \
  -p 8080:8080 \
  -v uptimelit-data:/data \
  ghcr.io/codewec/uptimelit:dev
```

To use a published image with Compose, replace `build: .` and `image: uptimelit:local` in `compose.yml` with `image: ghcr.io/codewec/uptimelit:dev`. Then update with:

```bash
docker compose pull
docker compose up -d
```

Monitors can reach other services on the same Docker network by service name. Inside the container, `localhost` refers to the UptimeLit container itself.

### Build from source

Requires Go 1.26.2 or later. From the project directory:

```bash
make build
./uptimelit
```

The binary listens on `:8080` and stores its database at `./uptimelit.db` by default.

## Users and sessions

Administrators can create users, change names and roles, reset passwords and OIDC bindings, and delete accounts together with their monitors, history and notifiers. Administrator privileges do not grant access to another user's monitoring data. The last administrator cannot be demoted or deleted, and users cannot delete their own accounts.

**Profile** lets a user change their username or password. Changing a password requires the current password when one exists. An OIDC-only user can set their first password from an OIDC session if password login is enabled. Usernames cannot be changed during an OIDC session.

Sessions last seven days. Signing out revokes the current session. Profile changes, administrative account updates and initial OIDC account linking revoke the affected user's existing sessions. Deleting a user revokes their sessions and stops their monitors.

## Configuration

Monitor and notifier settings live in the database and are managed through the UI or API. Server settings come from environment variables; `-addr` and `-db` override their corresponding environment values.

| Environment variable | Default | Purpose |
| --- | --- | --- |
| `UPTIMELIT_ADDR` | `:8080` | HTTP listen address; also settable with `-addr` |
| `UPTIMELIT_DB` | `./uptimelit.db` | SQLite path; also settable with `-db`. Docker defaults to `/data/uptimelit.db` |
| `JWT_SECRET` | Generated and stored in the database | Optional session signing key, at least 32 bytes |
| `OIDC_ISSUER` | Empty | OIDC provider issuer URL |
| `OIDC_CLIENT_ID` | Empty | OIDC client ID |
| `OIDC_CLIENT_SECRET` | Empty | OIDC client secret |
| `OIDC_REDIRECT_URL` | Empty | Required absolute callback URL when OIDC is enabled |
| `OIDC_BUTTON_TITLE` | `Sign in with OIDC` | OIDC login button label |
| `DISABLE_PASSWORD_LOGIN` | `false` | Disable password login when OIDC is configured |
| `OIDC_INSECURE_SKIP_TLS_VERIFY` | `false` | Disable TLS certificate verification for OIDC requests only |

The binary does not automatically load `.env` files. The included Compose file passes its listed authentication settings from the shell or a project `.env` file into the container. For `docker run`, use `-e` or `--env-file` to pass settings.

### OIDC

Register an OIDC client with your provider and set its callback to:

```text
https://monitor.example.com/api/auth/oidc/callback
```

For the included Compose service, create a `.env` file alongside `compose.yml`:

```dotenv
OIDC_ISSUER=https://id.example.com
OIDC_CLIENT_ID=uptimelit
OIDC_CLIENT_SECRET=provider-issued-secret
OIDC_REDIRECT_URL=https://monitor.example.com/api/auth/oidc/callback
OIDC_BUTTON_TITLE=Sign in with Pocket ID
```

Recreate the container with `docker compose up -d`. The issuer and client ID must both be supplied, and the callback must match the provider configuration. UptimeLit uses authorization code flow with PKCE, state and nonce validation, and verifies the ID token's signature, issuer, audience and expiry.

An existing OIDC binding is identified by issuer and subject. For an unbound identity, UptimeLit links to an administrator-created account with the same normalized username, preserving its password and role. The username comes from `preferred_username`, then the email local part, then display name; unsupported characters are replaced by dashes. Use a trusted provider whose usernames you control.

After bootstrap, unknown identities are denied: OIDC does not enable public registration. Administrators can create accounts without passwords when OIDC is configured, then users can sign in through the matching provider identity.

Password login remains available by default. Set `DISABLE_PASSWORD_LOGIN=true` after verifying OIDC if you want OIDC-only login. Without configured OIDC, password login remains enabled.

For HTTPS behind a reverse proxy, forward the original host and `X-Forwarded-Proto` header. Prefer a trusted CA for private providers. `OIDC_INSECURE_SKIP_TLS_VERIFY=true` explicitly bypasses certificate verification for OIDC discovery, token exchange and key retrieval, and emits a warning at startup.

## Monitoring and notifications

HTTP checks try `HEAD` first and retry with `GET` if the server returns `405 Method Not Allowed`. Status codes `200`–`399` count as UP; redirects are not followed. The HTTP client timeout is 10 seconds. TCP checks attempt to open a connection to the configured `host:port`.

In the UI, create notification channels in **Notifiers**, then select them when creating or editing a monitor. Supported channels are Telegram, Pushover and Redis. A monitor can only use its owner's notifiers. The failure threshold controls how many consecutive failed checks trigger an outage notification; recovery notifications are sent when the service comes back up.

## REST API

All monitoring, notifier and user-management routes require authentication. Sign in with a password:

```bash
curl -H 'Content-Type: application/json' -H 'X-UptimeLit-Request: 1' \
  -d '{"username":"alice","password":"your-password"}' \
  http://localhost:8080/api/auth/login
```

The response contains `token` and `user`. Pass the token as `Authorization: Bearer <token>` for subsequent requests. The browser uses an HttpOnly session cookie. All writes without a Bearer token, including login, require `X-UptimeLit-Request: 1` for CSRF protection.

### Authentication and accounts

| Method | Route | Purpose |
| --- | --- | --- |
| GET | `/api/auth/config` | Public login-method and bootstrap configuration |
| POST | `/api/auth/login` | Password sign-in; creates the first administrator on an empty database |
| POST | `/api/auth/logout` | Revoke the current session |
| GET | `/api/auth/me` | Current user |
| PUT | `/api/auth/profile` | Update `username`, optionally `newPassword` and `currentPassword` |
| GET | `/api/auth/oidc/login` | Start OIDC login |
| GET | `/api/auth/oidc/callback` | OIDC provider callback |
| GET / POST | `/api/admin/users` | List or create users; administrators only |
| PUT / DELETE | `/api/admin/users/{id}` | Update or delete users; administrators only |

User creation accepts `username`, `role` (`admin` or `user`, default `user`) and `password` (optional with OIDC). Updates accept `username`, `role`, optional `newPassword` and `resetOIDC`. Resetting an OIDC binding requires an available password login method.

### Monitors and notifiers

Every resource route is scoped to the signed-in user. Accessing another user's resource returns `404`, including for administrators.

| Method | Route | Purpose |
| --- | --- | --- |
| GET / POST | `/api/endpoints` | List own monitors with current statistics, or create a monitor |
| GET / PUT / DELETE | `/api/endpoints/{id}` | Read, update or delete a monitor and its history |
| POST | `/api/endpoints/{id}/pause` | Toggle monitoring pause |
| GET | `/api/endpoints/{id}/history?limit=60` | Recent history for the timeline |
| GET | `/api/endpoints/{id}/checks?limit=100` | Check details with status codes and response times |
| GET / POST | `/api/notifiers` | List own notifiers or create a notifier |
| GET / PUT / DELETE | `/api/notifiers/{id}` | Read, update or delete a notifier |
| POST | `/api/notifiers/{id}/test` | Send a test notification |
| GET | `/api/status` | Application version |

## Development

```bash
make test
go test -race ./...
go vet ./...
```

Database migrations live in [internal/db/migrations](internal/db/migrations/README.md) as timestamped Go files with apply and rollback functions. Startup applies pending migrations automatically, committing each schema change and its `schema_migrations` entry together. Back up the database before deploying schema changes.

Contributions are welcome through issues and pull requests.

## Acknowledgements

UptimeLit builds on the original [TinyPulse project](https://github.com/AkaCoder404/tinypulse). Thanks to its author and contributors for the foundation.

## License

[MIT](LICENSE). The original copyright notice is retained in the license file.
