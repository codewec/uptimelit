// Package migrations contains ordered, transactional SQLite schema migrations.
package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
)

type migration struct {
	version  int64
	name     string
	up, down func(context.Context, *sql.Tx) error
}

var registered []migration

// MustRegister is called from init in each timestamped migration file.
// Both callbacks run inside a transaction and must use the supplied tx.
func MustRegister(version int64, name string, up, down func(context.Context, *sql.Tx) error) {
	if version <= 0 || name == "" || up == nil || down == nil {
		panic("invalid migration registration")
	}
	for _, m := range registered {
		if m.version == version {
			panic(fmt.Sprintf("duplicate migration version %d", version))
		}
	}
	registered = append(registered, migration{version, name, up, down})
}

// Apply runs pending migrations in version order.
func Apply(ctx context.Context, conn *sql.DB) error { return apply(ctx, conn, registered) }

func apply(ctx context.Context, conn *sql.DB, all []migration) error {
	ordered := append([]migration(nil), all...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].version < ordered[j].version })
	for _, m := range ordered {
		if err := applyOne(ctx, conn, ordered, m); err != nil {
			return err
		}
	}
	return nil
}

func begin(ctx context.Context, conn *sql.DB, all []migration) (*sql.Tx, map[int64]bool, error) {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	applied, err := readLedger(ctx, tx, all)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	return tx, applied, nil
}

func readLedger(ctx context.Context, tx *sql.Tx, all []migration) (map[int64]bool, error) {
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
 version INTEGER PRIMARY KEY, name TEXT NOT NULL, applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
 )`); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT version,name FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	known := make(map[int64]string, len(all))
	for _, m := range all {
		known[m.version] = m.name
	}
	applied := make(map[int64]bool)
	for rows.Next() {
		var version int64
		var name string
		if err = rows.Scan(&version, &name); err != nil {
			return nil, err
		}
		expected, ok := known[version]
		if !ok {
			return nil, fmt.Errorf("database contains unknown migration %d; use a compatible application version", version)
		}
		if name != expected {
			return nil, fmt.Errorf("migration %d name mismatch: database %q, application %q", version, name, expected)
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func applyOne(ctx context.Context, conn *sql.DB, all []migration, m migration) error {
	tx, applied, err := begin(ctx, conn, all)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if applied[m.version] {
		return tx.Commit()
	}
	for version := range applied {
		if version > m.version {
			return fmt.Errorf("cannot apply migration %d before already applied migration %d", m.version, version)
		}
	}
	if err = m.up(ctx, tx); err != nil {
		return fmt.Errorf("apply migration %d_%s: %w", m.version, m.name, err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations(version,name) VALUES(?,?)`, m.version, m.name); err != nil {
		return err
	}
	return tx.Commit()
}

// Rollback reverts the last applied migration and removes its ledger entry in
// the same transaction. It is never invoked automatically at application startup.
// Rolling back the initial migration deletes all application data.
func Rollback(ctx context.Context, conn *sql.DB) error { return rollback(ctx, conn, registered) }
func rollback(ctx context.Context, conn *sql.DB, all []migration) error {
	tx, applied, err := begin(ctx, conn, all)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var last *migration
	for i := range all {
		m := &all[i]
		if applied[m.version] && (last == nil || m.version > last.version) {
			last = m
		}
	}
	if last == nil {
		return tx.Commit()
	}
	if err = last.down(ctx, tx); err != nil {
		return fmt.Errorf("rollback migration %d_%s: %w", last.version, last.name, err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version=?`, last.version); err != nil {
		return err
	}
	return tx.Commit()
}
