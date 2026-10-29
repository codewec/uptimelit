package db

import (
	"context"
	"database/sql"
	"net/url"
	"path/filepath"

	"uptimelit/internal/db/migrations"

	_ "modernc.org/sqlite"
)

type DB struct{ conn *sql.DB }
type ownerKey struct{}

func WithOwner(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ownerKey{}, id)
}
func Owner(ctx context.Context) string { v, _ := ctx.Value(ownerKey{}).(string); return v }

// SQL is for trusted authentication and maintenance code, never user-supplied queries.
func (d *DB) SQL() *sql.DB { return d.conn }
func New(path string) (*DB, error) {
	dsn := path
	if path != ":memory:" {
		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		dsn = (&url.URL{Scheme: "file", Path: abs}).String()
	}
	conn, err := sql.Open("sqlite", dsn+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_txlock=immediate")
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	d := &DB{conn: conn}
	if err = d.migrate(); err != nil {
		conn.Close()
		return nil, err
	}
	return d, nil
}
func (d *DB) Close() error { return d.conn.Close() }
func (d *DB) migrate() error {
	if _, err := d.conn.Exec(`PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;`); err != nil {
		return err
	}
	return migrations.Apply(context.Background(), d.conn)
}
