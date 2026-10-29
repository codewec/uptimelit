package migrations

import (
	"context"
	"database/sql"
)

func init() {
	MustRegister(20261029000001, "init", func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
 CREATE TABLE users (
 id TEXT PRIMARY KEY, username TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL DEFAULT '',
 role TEXT NOT NULL CHECK(role IN ('admin','user')), oidc_issuer TEXT, oidc_subject TEXT,
 UNIQUE(oidc_issuer,oidc_subject));
 CREATE TABLE settings (key TEXT PRIMARY KEY,value TEXT NOT NULL);
 CREATE TABLE sessions (id TEXT PRIMARY KEY,user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,auth_method TEXT NOT NULL,expires_at INTEGER NOT NULL);
 CREATE INDEX idx_sessions_user ON sessions(user_id);
 CREATE TABLE endpoints (
 id INTEGER PRIMARY KEY AUTOINCREMENT,owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 type TEXT NOT NULL DEFAULT 'http',name TEXT NOT NULL,url TEXT NOT NULL,interval_seconds INTEGER NOT NULL DEFAULT 60,
 fail_threshold INTEGER NOT NULL DEFAULT 3,paused INTEGER NOT NULL DEFAULT 0,created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
 CREATE INDEX idx_endpoints_owner ON endpoints(owner_id);
 CREATE TABLE notifiers (
 id INTEGER PRIMARY KEY AUTOINCREMENT,owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 name TEXT NOT NULL,type TEXT NOT NULL,config_json TEXT NOT NULL,created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
 CREATE INDEX idx_notifiers_owner ON notifiers(owner_id);
 CREATE TABLE checks (
 id INTEGER PRIMARY KEY AUTOINCREMENT,endpoint_id INTEGER NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
 status_code INTEGER,response_time_ms INTEGER NOT NULL,is_up INTEGER NOT NULL,checked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 error_code TEXT NOT NULL DEFAULT '');
 CREATE INDEX idx_checks_endpoint_checked_up ON checks(endpoint_id,checked_at,is_up);
 CREATE INDEX idx_checks_checked_at ON checks(checked_at);
 CREATE TABLE endpoint_notifiers (
 endpoint_id INTEGER NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
 notifier_id INTEGER NOT NULL REFERENCES notifiers(id) ON DELETE CASCADE,PRIMARY KEY(endpoint_id,notifier_id));
 CREATE TRIGGER same_owner_insert BEFORE INSERT ON endpoint_notifiers
 WHEN (SELECT owner_id FROM endpoints WHERE id=NEW.endpoint_id) IS NOT (SELECT owner_id FROM notifiers WHERE id=NEW.notifier_id)
 BEGIN SELECT RAISE(ABORT,'notifier does not belong to endpoint owner'); END;
`)
		return err
	}, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
DROP TRIGGER same_owner_insert;
DROP TABLE endpoint_notifiers;
DROP TABLE checks;
DROP TABLE notifiers;
DROP TABLE endpoints;
DROP TABLE sessions;
DROP TABLE settings;
DROP TABLE users;
`)
		return err
	})
}
