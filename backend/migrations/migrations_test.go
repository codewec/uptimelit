package migrations

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"

	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)&_txlock=immediate")
	if err != nil {
		t.Fatal(err)
	}
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { conn.Close() })
	return conn
}

func execSQL(query string) func(context.Context, *sql.Tx) error {
	return func(ctx context.Context, tx *sql.Tx) error { _, err := tx.ExecContext(ctx, query); return err }
}

func TestApplyOrdersAndSkipsCompletedMigrations(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	// These statements cannot be run twice and must run in version order.
	all := []migration{
		{2, "value", execSQL(`ALTER TABLE items ADD COLUMN value TEXT NOT NULL DEFAULT 'kept'`), execSQL(`ALTER TABLE items DROP COLUMN value`)},
		{1, "items", execSQL(`CREATE TABLE items(id INTEGER PRIMARY KEY);INSERT INTO items(id) VALUES(1)`), execSQL(`DROP TABLE items`)},
	}
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := apply(ctx, conn, all); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	var value string
	if err := conn.QueryRow(`SELECT value FROM items WHERE id=1`).Scan(&value); err != nil || value != "kept" {
		t.Fatal(value, err)
	}
	var count int
	if err := conn.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&count); err != nil || count != 2 {
		t.Fatal(count, err)
	}
	if err := rollback(ctx, conn, all); err != nil {
		t.Fatal(err)
	}
	if err := conn.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&count); err != nil || count != 1 {
		t.Fatal(count, err)
	}
	if err := apply(ctx, conn, all); err != nil {
		t.Fatal(err)
	}
	if err := conn.QueryRow(`SELECT value FROM items WHERE id=1`).Scan(&value); err != nil || value != "kept" {
		t.Fatal(value, err)
	}
}

func TestFailedMigrationRollsBackSchemaAndLedger(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	all := []migration{
		{1, "items", execSQL(`CREATE TABLE items(id INTEGER)`), execSQL(`DROP TABLE items`)},
		{2, "broken", func(ctx context.Context, tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, `ALTER TABLE items ADD COLUMN partial TEXT`); err != nil {
				return err
			}
			return errors.New("deliberate failure")
		}, execSQL(`SELECT 1`)},
	}
	if err := apply(ctx, conn, all); err == nil {
		t.Fatal("failed migration accepted")
	}
	var count int
	if err := conn.QueryRow(`SELECT count(*) FROM pragma_table_info('items') WHERE name='partial'`).Scan(&count); err != nil || count != 0 {
		t.Fatal("partial schema persisted", count, err)
	}
	if err := conn.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&count); err != nil || count != 1 {
		t.Fatal("incorrect ledger", count, err)
	}
	// A failed rollback also restores schema and preserves its ledger entry.
	all[0].down = func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DROP TABLE items`); err != nil {
			return err
		}
		return errors.New("deliberate rollback failure")
	}
	if err := rollback(ctx, conn, all); err == nil {
		t.Fatal("failed rollback accepted")
	}
	if err := conn.QueryRow(`SELECT count(*) FROM items`).Scan(&count); err != nil {
		t.Fatal("rollback lost table", err)
	}
	if err := conn.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&count); err != nil || count != 1 {
		t.Fatal(count, err)
	}
}

func TestUnknownMigrationRejected(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	all := []migration{{2, "future", execSQL(`CREATE TABLE future(id INTEGER)`), execSQL(`DROP TABLE future`)}}
	if err := apply(ctx, conn, all); err != nil {
		t.Fatal(err)
	}
	older := []migration{{1, "old", execSQL(`CREATE TABLE old(id INTEGER)`), execSQL(`DROP TABLE old`)}}
	if err := apply(ctx, conn, older); err == nil {
		t.Fatal("unknown migration accepted")
	}
	var count int
	if err := conn.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name='old'`).Scan(&count); err != nil || count != 0 {
		t.Fatal("old schema applied", count, err)
	}
}

func TestInitialMigrationRollbackAndReapply(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	if err := Apply(ctx, conn); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`
 INSERT INTO users(id,username,role) VALUES('owner','alice','admin');
 INSERT INTO endpoints(id,owner_id,name,url) VALUES(1,'owner','site','https://example.com');
 INSERT INTO notifiers(id,owner_id,name,type,config_json) VALUES(1,'owner','alerts','telegram','{}');
 INSERT INTO endpoint_notifiers VALUES(1,1);
 INSERT INTO sessions VALUES('session','owner','password',9999999999);
 INSERT INTO checks(endpoint_id,response_time_ms,is_up) VALUES(1,1,1);
 `); err != nil {
		t.Fatal(err)
	}
	if err := Rollback(ctx, conn); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := conn.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name IN ('users','endpoints','notifiers','sessions','checks','settings','endpoint_notifiers','same_owner_insert')`).Scan(&count); err != nil || count != 0 {
		t.Fatal("initial schema not removed", count, err)
	}
	if err := conn.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&count); err != nil || count != 0 {
		t.Fatal("ledger not cleared", count, err)
	}
	if err := Apply(ctx, conn); err != nil {
		t.Fatal(err)
	}
	if err := conn.QueryRow(`SELECT count(*) FROM users`).Scan(&count); err != nil || count != 0 {
		t.Fatal("schema not recreated", count, err)
	}
}
