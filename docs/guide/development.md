# Development

## Requirements and structure

Use Go 1.26.2+, Node.js 22+ and pnpm 10.15.0. The root is an orchestration directory; the Go module and frontend package are separate:

```text
backend/
  cmd/server/          # Entry point and generated embedded SPA files
  internal/auth/       # Password, session and OIDC services
  internal/db/         # SQLite access
  internal/model/      # Monitoring models
  internal/monitor/    # Workers and cached statistics
  internal/notifier/   # Alert providers
  internal/server/     # HTTP API and SPA routing
  migrations/          # Timestamped Go migrations
frontend/
  src/components/      # Shared Svelte UI and dialogs
  src/layouts/         # App and authentication shells
  src/pages/           # Monitors, notifiers, login, profile and administration
  src/lib/             # Typed API, router, state and themes
  public/              # Public assets
  tests/               # Browser workflows and isolated test server
```

The frontend uses Svelte 5, TypeScript, Vite, Tailwind CSS 4 and Bits UI. Its palettes, Inter fonts, cards, menus and dialogs follow DashLit's visual system. Fonts and styles are bundled locally. See [frontend attribution](../../frontend/THIRD_PARTY.md).

## Separate development servers

```bash
cd frontend
pnpm install --frozen-lockfile
cd ..
make dev-backend
```

In another terminal:

```bash
make dev-frontend
```

Open `http://localhost:5173`. The backend listens on port `8080`; Vite proxies `/api` there while preserving the original host for same-origin cookie/CSRF checks. By default `make dev-backend` runs from `backend/` and uses `backend/uptimelit.db`. Export `UPTIMELIT_DB` to choose another path.

The backend can serve the API before the frontend is built. Page requests to an API-only build report that frontend assets are missing; use Vite during development or create a production build.

## Production build

```bash
make build
./app
```

The Makefile builds `frontend/dist`, copies it into `backend/cmd/server/static` and compiles `./app`. Generated assets and the binary are not committed. The Dockerfile performs the same frontend/backend stages and runs the embedded binary as a non-root user.

`make backend` compiles the currently available embedded files without rebuilding the frontend. Use `make build` when you need current UI assets. `make clean` removes build output and preserves databases.

## Checks and tests

```bash
make check
make test
```

`make check` runs Svelte/TypeScript checks and `go vet`. `make test` runs backend tests with the race detector. Direct Go commands belong in the module directory, for example `cd backend && go test -race ./...`.

To run browser workflows, install Chromium once:

```bash
cd frontend
pnpm exec playwright install chromium
cd ..
make test-e2e
```

The suite builds the app and starts its own server on `127.0.0.1:18093` with a temporary database. It covers bootstrap/login, status badge filters and search reset, navbar creation, monitor/notifier CRUD, history, user isolation, profile sessions and mobile navigation. Its notification-send action is intercepted rather than delivering to an external provider.

## Migrations

Add schema changes under [backend/migrations](../../backend/migrations/README.md), using a timestamped Go filename and registered up/down callbacks. Each migration and its ledger entry commit together. Startup applies only pending migrations and rejects unknown recorded versions. See the migration README for transaction and rollback rules.

## Documentation and screenshots

The [documentation index](../index.md) links to the user guides. Screenshots live in `.github/images/` and are referenced from both the README and [gallery](screenshots.md). Keep their navbar and filter controls aligned with the current UI when refreshing them.
