package db

import (
	"path/filepath"
	"testing"
)

func TestDatabaseReopenPreservesData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "users.db")
	d, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	// Reopening the current schema must preserve user data and live sessions.
	_, err = d.SQL().Exec(`
 INSERT INTO users(id,username,password_hash,role) VALUES('owner','alice','existing-hash','admin');
 INSERT INTO settings(key,value) VALUES('jwt_secret','existing-secret');
 INSERT INTO sessions VALUES('session','owner','password',9999999999);
 INSERT INTO endpoints(id,owner_id,type,name,url) VALUES(1,'owner','http','monitor','https://example.com');
 INSERT INTO notifiers(id,owner_id,name,type,config_json) VALUES(1,'owner','alerts','telegram','{}');
 INSERT INTO endpoint_notifiers VALUES(1,1);
 INSERT INTO checks(endpoint_id,response_time_ms,is_up) VALUES(1,5,1);
 `)
	if err != nil {
		d.Close()
		t.Fatal(err)
	}
	d.Close()
	d, err = New(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"users", "settings", "sessions", "endpoints", "notifiers", "endpoint_notifiers", "checks", "schema_migrations"} {
		var count int
		if err = d.SQL().QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 1 {
			d.Close()
			t.Fatalf("%s count=%d error=%v", table, count, err)
		}
	}
	var secret string
	if err = d.SQL().QueryRow(`SELECT value FROM settings WHERE key='jwt_secret'`).Scan(&secret); err != nil || secret != "existing-secret" {
		d.Close()
		t.Fatal(secret, err)
	}
	// A subsequent startup must leave the completed migration record untouched.
	if _, err = d.SQL().Exec(`UPDATE schema_migrations SET applied_at='2000-01-01 00:00:00'`); err != nil {
		d.Close()
		t.Fatal(err)
	}
	d.Close()
	d, err = New(path)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	var completed string
	if err = d.SQL().QueryRow(`SELECT CAST(applied_at AS TEXT) FROM schema_migrations`).Scan(&completed); err != nil || completed != "2000-01-01 00:00:00" {
		t.Fatal(completed, err)
	}
}
