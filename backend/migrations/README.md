# Database migrations

Migrations are compiled into the binary and applied automatically by `db.New`.
No external SQL files or additional Docker copies are needed.

Add a Go file named `YYYYMMDDHHMMSS_description.go` in this package. In `init`,
call `MustRegister` with the matching numeric timestamp, a stable name, and
`up`/`down` callbacks accepting `context.Context` and `*sql.Tx`. Use only the
supplied transaction for database operations. Do not execute transaction control
or connection-level PRAGMA statements inside a migration.

Use a new, increasing timestamp for every schema change. Never edit or rename
an applied migration. Migrations run in version order; each change and its
`schema_migrations` entry commit together. A failed migration rolls back and
startup reports the error. Completed earlier migrations remain committed.
Unknown versions, renamed entries and pending migrations older than completed
versions are rejected rather than silently modifying an incompatible database.

The initial migration creates a new database schema. Existing installations must
have a matching `schema_migrations` ledger; unversioned schemas are not adopted.

`Rollback` explicitly reverts the most recent migration and removes its ledger
entry in one transaction. Startup never invokes it. Rolling back the initial
migration deletes all application data; this function has no HTTP or CLI route.
