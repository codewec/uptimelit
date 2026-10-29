# Backups and restore

## What to preserve

The SQLite database contains users and password hashes, OIDC bindings, monitors, notifiers, check history, sessions and the generated session signing key. For the standard container, the files live under `/data` and the main database is `/data/uptimelit.db`.

Also preserve deployment environment settings, including OIDC credentials and `JWT_SECRET` when explicitly configured. Theme preferences are stored in each browser and are not part of the database.

A live SQLite database may have `uptimelit.db-wal` and `uptimelit.db-shm` files. Stop the service before making a filesystem copy of the whole data directory, or use SQLite's online backup tooling to obtain a consistent snapshot. Copying only a live `.db` file can omit committed data still in its WAL.

## Compose volume backup

The following example uses the stopped service container's volume, so it does not depend on the Compose project prefix:

```bash
backup_dir="$HOME/uptimelit-backups"
mkdir -p "$backup_dir"
docker compose stop uptimelit
docker run --rm \
  --volumes-from "$(docker compose ps -aq uptimelit)" \
  -v "$backup_dir:/backup" \
  alpine:3.23 tar -czf /backup/uptimelit-data.tar.gz -C /data .
docker compose start uptimelit
```

Use a distinct archive name for each backup you want to retain. Run this from the directory containing your deployment's Compose file. The service must have been created already so that `docker compose ps -aq uptimelit` can find its container.

## Restore

1. Stop UptimeLit and retain the current volume until the restore is verified.
2. Restore the archive into an empty replacement volume or data directory. Do not mix a database snapshot with WAL files from another database state.
3. Preserve UID/GID `10001:10001` for the container data files and reconnect the replacement storage to `/data`.
4. Restore the deployment environment, including an explicit signing key if one was used.
5. Start UptimeLit, inspect logs and verify login, monitors, history and notification settings.

For native deployments, follow the same process for the directory containing the configured `UPTIMELIT_DB` path.

## Updates and schema changes

Back up before updating. Each migration commits with its `schema_migrations` entry in one transaction. A failed migration rolls back that migration; earlier completed migrations remain applied. Downgrading an executable does not downgrade the database.

If restoring an older backup, use an application version that supports its schema. Developer rollback functions are not a substitute for a data backup. See [installation and updates](installation.md) and [migration development](../../backend/migrations/README.md).
