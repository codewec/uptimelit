# Installation and updates

## Run the published Docker image

```bash
docker pull ghcr.io/codewec/uptimelit:dev
docker run -d \
  --name uptimelit \
  --restart unless-stopped \
  -p 8080:8080 \
  -v uptimelit-data:/data \
  ghcr.io/codewec/uptimelit:dev
```

Open `http://localhost:8080` and [create the first administrator](getting-started.md). The `dev` tag is updated by the [workflow](../../.github/workflows/docker-dev.yml) on pushes to `main`; it is a development channel.

The image runs as UID/GID `10001:10001`, stores its database at `/data/uptimelit.db` and serves both the UI and API on port `8080`. Its healthcheck requests `/api/auth/config`.

## Docker Compose from source

```bash
git clone https://github.com/codewec/uptimelit.git
cd uptimelit
docker compose up -d --build
```

The included [compose.yml](../../compose.yml) builds `uptimelit:local`. Its `uptimelit-data` named volume persists the database and WAL files. Compose normally prefixes the actual volume name with the project name.

For authentication options, copy [.env.example](../../.env.example) to `.env`, fill in the required settings and recreate the service. OIDC is optional; see [configuration](configuration.md).

## Docker Compose with the published image

In `compose.yml`, remove `build: .` and replace `image: uptimelit:local` with:

```yaml
image: ghcr.io/codewec/uptimelit:dev
```

Keep the existing environment and volume settings, then run:

```bash
docker compose pull
docker compose up -d
```

## Updates

Back up the database before deploying schema changes. Migrations run automatically at startup and commit individually in transactions; a failed migration stops startup. Older binaries may not understand a newer schema. See [backups and restore](backups.md).

For an image built from the checkout:

```bash
git pull
docker compose up -d --build
docker compose logs --tail=100 uptimelit
```

For the published image:

```bash
docker compose pull
docker compose up -d
docker compose logs --tail=100 uptimelit
```

For a standalone Docker container, pull the new image, stop and remove the old container, then recreate it with the same volume and environment settings. Removing the container does not remove the named volume.

`docker compose down` preserves the volume. `docker compose down -v` deletes it.

## Build a standalone binary

Requires Go 1.27.3+, Node.js 26.13+ and pnpm 12.16.1:

```bash
make build
./app
```

The build embeds the Svelte app and its fonts and styles into `./app`; Node.js is not needed at runtime. By default the binary listens on `:8080` and uses `./uptimelit.db` relative to its working directory. See [development](development.md) for separate development servers and build commands.

## Networking

Checks run from the UptimeLit server/container. On the same Docker network, monitor another service by its service name. `localhost` inside the container means the UptimeLit container itself.

For a host bind mount instead of a named volume, make the data directory writable by UID/GID `10001:10001`. For HTTPS, put a reverse proxy in front of the service and forward the original host and `X-Forwarded-Proto` headers.
