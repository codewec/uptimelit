package auth

import (
	"context"
	"github.com/codewec/uptimelit/backend/internal/db"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func service(t *testing.T) (*Service, *db.DB) {
	t.Helper()
	d, e := db.New(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { d.Close() })
	s, e := New(d, Config{})
	if e != nil {
		t.Fatal(e)
	}
	return s, d
}
func TestConcurrentBootstrap(t *testing.T) {
	s, _ := service(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	start := make(chan struct{})
	for _, name := range []string{"one", "two", "three", "four"} {
		wg.Add(1)
		go func(n string) { defer wg.Done(); <-start; s.Login(ctx, n, "strong-password") }(name)
	}
	close(start)
	wg.Wait()
	users, e := s.Users(ctx)
	if e != nil || len(users) != 1 || users[0].Role != "admin" {
		t.Fatalf("bootstrap users=%+v error=%v", users, e)
	}
}
func TestOIDCProvisioningAndBinding(t *testing.T) {
	s, _ := service(t)
	ctx := context.Background()
	i := &OIDCIdentity{Issuer: "https://issuer", Subject: "first", PreferredUsername: "root"}
	u, e := s.OIDCUser(ctx, i)
	if e != nil || u.Role != "admin" {
		t.Fatalf("bootstrap: %+v %v", u, e)
	}
	if _, e = s.OIDCUser(ctx, &OIDCIdentity{Issuer: i.Issuer, Subject: "unknown", PreferredUsername: "new-user"}); e != ErrDenied {
		t.Fatalf("unexpected provisioning: %v", e)
	}
	local, e := s.CreateUser(ctx, "alice", "alice-password", "user")
	if e != nil {
		t.Fatal(e)
	}
	identity := &OIDCIdentity{Issuer: i.Issuer, Subject: "alice-sub", PreferredUsername: "alice"}
	linked, e := s.OIDCUser(ctx, identity)
	if e != nil || linked.ID != local.ID || !linked.HasPassword {
		t.Fatalf("link: %+v %v", linked, e)
	}
	identity.Subject = "different"
	if _, e = s.OIDCUser(ctx, identity); e != ErrDenied {
		t.Fatalf("rebound identity: %v", e)
	}
	if e = s.UpdateUser(ctx, u.ID, "root", "admin", "", true, true); e == nil {
		t.Fatal("removed only login method")
	}
	if e = s.UpdateUser(ctx, u.ID, "root", "admin", "new-password", true, true); e != nil {
		t.Fatal(e)
	}
}
func TestSessionsPersistExpireAndRevoke(t *testing.T) {
	path := filepath.Join(t.TempDir(), "users.db")
	d, e := db.New(path)
	if e != nil {
		t.Fatal(e)
	}
	s, e := New(d, Config{})
	if e != nil {
		t.Fatal(e)
	}
	ctx := context.Background()
	u, e := s.Login(ctx, "root", "admin-password")
	if e != nil {
		t.Fatal(e)
	}
	token, e := s.Session(ctx, u, "password")
	if e != nil {
		t.Fatal(e)
	}
	d.Close()
	d, e = db.New(path)
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	s, e = New(d, Config{})
	if e != nil {
		t.Fatal(e)
	}
	if _, _, e = s.Authenticate(ctx, token); e != nil {
		t.Fatal("session lost on restart", e)
	}
	if _, e = d.SQL().Exec(`UPDATE sessions SET expires_at=?`, time.Now().Add(-time.Second).Unix()); e != nil {
		t.Fatal(e)
	}
	if _, _, e = s.Authenticate(ctx, token); e == nil {
		t.Fatal("expired session accepted")
	}
	token, e = s.Session(ctx, u, "password")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.UpdateUser(ctx, u.ID, u.Username, u.Role, "updated-password", false, true); e != nil {
		t.Fatal(e)
	}
	if _, _, e = s.Authenticate(ctx, token); e == nil {
		t.Fatal("revoked session accepted")
	}
}
func TestPasswordLoginConfig(t *testing.T) {
	if !(Config{DisablePasswordLogin: true}).PasswordLoginEnabled() {
		t.Fatal("password fallback disabled without OIDC")
	}
	if (Config{DisablePasswordLogin: true, OIDCIssuer: "https://issuer", OIDCClientID: "client"}).PasswordLoginEnabled() {
		t.Fatal("OIDC-only mode ignored")
	}
}
