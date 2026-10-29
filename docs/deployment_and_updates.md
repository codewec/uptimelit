# Deploying and Updating UptimeLit

## Deployment with Docker Compose

From the project directory, build and start the included Compose service:

```bash
docker compose up -d --build
```

Open `http://localhost:8080`. The first sign-in creates the administrator.
Configure OIDC through the Compose environment if needed; see the
[authentication settings](../README.md#authentication-and-configuration).

The `uptimelit-data` named volume stores the database in `/data/uptimelit.db`.
It survives container recreation. `docker compose down` preserves the volume;
`docker compose down -v` deletes it.

For HTTPS, place a reverse proxy in front of port 8080 and forward the original
host and `X-Forwarded-Proto` header.

## Updating

Back up the database volume with the service stopped before deploying schema
changes. Then obtain the updated source and rebuild the container:

```bash
git pull
docker compose up -d --build
docker compose logs --tail=100 uptimelit
```

## Database migrations

Versioned migrations run automatically at startup. Each migration and its entry
in `schema_migrations` commit in one transaction. A failed migration is rolled
back and startup stops with an error.

For adding migrations, see [the migration guide](../internal/db/migrations/README.md).
