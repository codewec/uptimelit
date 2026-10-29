package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uptimelit/internal/auth"
	"uptimelit/internal/db"
	"uptimelit/internal/model"
	"uptimelit/internal/monitor"
	"uptimelit/internal/notifier"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	d, e := db.New(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	dispatch := notifier.NewDispatcher(d)
	m := monitor.New(d, dispatch)
	ctx, cancel := context.WithCancel(context.Background())
	if e = m.Start(ctx); e != nil {
		t.Fatal(e)
	}
	s, e := New(d, m, auth.Config{})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { cancel(); m.Stop(); dispatch.Stop(); d.Close() })
	return s
}
func request(t *testing.T, s *Server, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	r := httptest.NewRequest(method, path, bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-UptimeLit-Request", "1")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}
func expect(t *testing.T, w *httptest.ResponseRecorder, code int) {
	t.Helper()
	if w.Code != code {
		t.Fatalf("status %d, want %d: %s", w.Code, code, w.Body.String())
	}
}
func loginToken(t *testing.T, s *Server, name, password string) string {
	t.Helper()
	w := request(t, s, "POST", "/api/auth/login", "", map[string]string{"username": name, "password": password})
	expect(t, w, 200)
	var v struct{ Token string }
	if e := json.Unmarshal(w.Body.Bytes(), &v); e != nil {
		t.Fatal(e)
	}
	return v.Token
}
func TestAuthenticationAndUserManagement(t *testing.T) {
	s := testServer(t)
	ctx := context.Background()
	expect(t, request(t, s, "GET", "/api/endpoints", "", nil), 401)
	admin := loginToken(t, s, "root", "admin-password")
	expect(t, request(t, s, "POST", "/api/auth/login", "", map[string]string{"username": "stranger", "password": "long-password"}), 401)
	expect(t, request(t, s, "POST", "/api/admin/users", admin, map[string]string{"username": "alice", "password": "alice-password", "role": "user"}), 201)
	alice := loginToken(t, s, "alice", "alice-password")
	expect(t, request(t, s, "GET", "/api/admin/users", alice, nil), 403)
	users, e := s.auth.Users(ctx)
	if e != nil {
		t.Fatal(e)
	}
	var root, user *auth.User
	for _, u := range users {
		if u.Role == "admin" {
			root = u
		} else {
			user = u
		}
	}
	expect(t, request(t, s, "PUT", "/api/admin/users/"+root.ID, admin, map[string]string{"username": "root", "role": "user"}), 400)
	expect(t, request(t, s, "DELETE", "/api/admin/users/"+root.ID, admin, nil), 400)
	// Cookie writes need CSRF protection; same-origin headers are accepted.
	r := httptest.NewRequest("POST", "/api/auth/logout", nil)
	r.AddCookie(&http.Cookie{Name: "token", Value: alice})
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	expect(t, w, 403)
	r = httptest.NewRequest("POST", "/api/auth/logout", nil)
	r.AddCookie(&http.Cookie{Name: "token", Value: alice})
	r.Header.Set("X-UptimeLit-Request", "1")
	r.Header.Set("Origin", "https://evil.example")
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	expect(t, w, 403)
	expect(t, request(t, s, "POST", "/api/auth/logout", alice, nil), 200)
	expect(t, request(t, s, "GET", "/api/auth/me", alice, nil), 401)
	alice = loginToken(t, s, "alice", "alice-password")
	expect(t, request(t, s, "PUT", "/api/admin/users/"+user.ID, admin, map[string]string{"username": "alice", "role": "user", "newPassword": "reset-password"}), 200)
	expect(t, request(t, s, "GET", "/api/auth/me", alice, nil), 401)
	alice = loginToken(t, s, "alice", "reset-password")
	expect(t, request(t, s, "PUT", "/api/auth/profile", alice, map[string]string{"username": "alice", "newPassword": "profile-password", "currentPassword": "wrong"}), 403)
	expect(t, request(t, s, "PUT", "/api/auth/profile", alice, map[string]string{"username": "alice", "newPassword": "profile-password", "currentPassword": "reset-password"}), 200)
	expect(t, request(t, s, "GET", "/api/auth/me", alice, nil), 401)
	loginToken(t, s, "alice", "profile-password")
	expect(t, request(t, s, "DELETE", "/api/admin/users/"+user.ID, admin, nil), 204)
}
func TestOwnerIsolationAndAtomicLinks(t *testing.T) {
	s := testServer(t)
	admin := loginToken(t, s, "root", "admin-password")
	u, e := s.auth.CreateUser(context.Background(), "alice", "alice-password", "user")
	if e != nil {
		t.Fatal(e)
	}
	alice := loginToken(t, s, "alice", "alice-password")
	body := map[string]any{"name": "private", "type": "telegram", "config_json": `{"bot_token":"secret","chat_id":"123"}`, "owner_id": "forged"}
	w := request(t, s, "POST", "/api/notifiers", alice, body)
	expect(t, w, 201)
	var n model.Notifier
	json.Unmarshal(w.Body.Bytes(), &n)
	if n.OwnerID != u.ID {
		t.Fatal("owner must be server-assigned")
	}
	ep := map[string]any{"name": "private monitor", "type": "http", "url": "http://127.0.0.1:1", "interval_seconds": 60, "fail_threshold": 3, "notifier_ids": []int64{n.ID}}
	w = request(t, s, "POST", "/api/endpoints", alice, ep)
	expect(t, w, 201)
	var monitor model.Endpoint
	json.Unmarshal(w.Body.Bytes(), &monitor)
	for _, path := range []string{fmt.Sprint("/api/endpoints/", monitor.ID), fmt.Sprintf("/api/endpoints/%d/checks", monitor.ID), fmt.Sprintf("/api/endpoints/%d/history", monitor.ID), fmt.Sprint("/api/notifiers/", n.ID)} {
		expect(t, request(t, s, "GET", path, admin, nil), 404)
	}
	for _, path := range []string{fmt.Sprintf("/api/endpoints/%d/pause", monitor.ID), fmt.Sprintf("/api/notifiers/%d/test", n.ID)} {
		expect(t, request(t, s, "POST", path, admin, nil), 404)
	}
	for _, kind := range []string{"endpoints", "notifiers"} {
		id := n.ID
		if kind == "endpoints" {
			id = monitor.ID
		}
		for _, method := range []string{"PUT", "DELETE"} {
			expect(t, request(t, s, method, fmt.Sprintf("/api/%s/%d", kind, id), admin, ep), 404)
		}
		w = request(t, s, "GET", "/api/"+kind, admin, nil)
		expect(t, w, 200)
		if bytes.Contains(w.Body.Bytes(), []byte("private")) || bytes.Contains(w.Body.Bytes(), []byte("secret")) {
			t.Fatal("private data leaked")
		}
	}
	// The administrator cannot link Alice's notifier; no partial monitor is inserted.
	expect(t, request(t, s, "POST", "/api/endpoints", admin, ep), 400)
	w = request(t, s, "GET", "/api/endpoints", admin, nil)
	expect(t, w, 200)
	if w.Body.String() != "[]\n" {
		t.Fatal(w.Body.String())
	}
	// Invalid updates preserve both monitor attributes and its original notifier links.
	ep["name"] = "should roll back"
	ep["notifier_ids"] = []int64{99999}
	expect(t, request(t, s, "PUT", fmt.Sprint("/api/endpoints/", monitor.ID), alice, ep), 400)
	found, e := s.db.GetEndpoint(db.WithOwner(context.Background(), u.ID), monitor.ID)
	if e != nil || found.Name != "private monitor" {
		t.Fatalf("partial update: %+v %v", found, e)
	}
	expect(t, request(t, s, "DELETE", "/api/admin/users/"+u.ID, admin, nil), 204)
	expect(t, request(t, s, "GET", "/api/auth/me", alice, nil), 401)
	for _, table := range []string{"endpoints", "notifiers", "endpoint_notifiers", "checks"} {
		var count int
		if e = s.db.SQL().QueryRow("SELECT count(*) FROM " + table).Scan(&count); e != nil || count != 0 {
			t.Fatalf("leftover %s: %d %v", table, count, e)
		}
	}
	// A queued result arriving after deletion cannot recreate history.
	if e = s.db.RecordCheck(context.Background(), &model.Check{EndpointID: monitor.ID, CheckedAt: time.Now()}); e != nil {
		t.Fatal(e)
	}
}
func TestOIDCCallbackRejectsMissingState(t *testing.T) {
	s := testServer(t)
	expect(t, request(t, s, "GET", "/api/auth/oidc/callback?code=fake", "", nil), 404)
}

type callbackOIDC struct {
	called  bool
	invalid bool
}

func (f *callbackOIDC) AuthorizationURL(state, nonce, verifier string) string {
	return "https://id.example/authorize?state=" + state
}
func (f *callbackOIDC) Authenticate(_ context.Context, code, nonce, verifier string) (*auth.OIDCIdentity, error) {
	f.called = true
	if f.invalid || code != "good-code" || nonce != "nonce" || verifier != "verifier" {
		return nil, fmt.Errorf("invalid OIDC exchange")
	}
	return &auth.OIDCIdentity{Issuer: "https://id.example", Subject: "first", PreferredUsername: "oidc-root"}, nil
}
func TestOIDCCallbackFlow(t *testing.T) {
	s := testServer(t)
	fake := &callbackOIDC{}
	s.oidc = fake
	s.authConfig.OIDCRedirectURL = "https://app.example/api/auth/oidc/callback"
	callback := func(query string, withCookies bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "/api/auth/oidc/callback?"+query, nil)
		if withCookies {
			for k, v := range map[string]string{"oidc_state": "state", "oidc_nonce": "nonce", "oidc_verifier": "verifier"} {
				r.AddCookie(&http.Cookie{Name: k, Value: v})
			}
		}
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		return w
	}
	for _, q := range []string{"state=wrong&code=good-code", "state=state&error=access_denied", "state=state"} {
		w := callback(q, true)
		expect(t, w, 302)
		if !bytes.Contains([]byte(w.Header().Get("Location")), []byte("oidc_error=")) || fake.called {
			t.Fatal("invalid callback accepted")
		}
	}
	w := callback("state=state&code=good-code", false)
	expect(t, w, 302)
	if fake.called {
		t.Fatal("callback accepted without browser cookies")
	}
	fake.invalid = true
	w = callback("state=state&code=good-code", true)
	expect(t, w, 302)
	empty, e := s.auth.Empty(context.Background())
	if e != nil || !empty {
		t.Fatal("account created from invalid token")
	}
	fake.invalid = false
	w = callback("state=state&code=good-code", true)
	expect(t, w, 302)
	if w.Header().Get("Location") != "/" {
		t.Fatal(w.Header())
	}
	var token string
	cleared := 0
	for _, c := range w.Result().Cookies() {
		if c.Name == "token" {
			token = c.Value
			if !c.Secure || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
				t.Fatal("unsafe session cookie")
			}
		} else if c.MaxAge < 0 {
			cleared++
		}
	}
	if token == "" || cleared != 3 {
		t.Fatal("missing session or uncleared flow cookies")
	}
	expect(t, request(t, s, "GET", "/api/auth/me", token, nil), 200)
}

func TestLoginRateLimit(t *testing.T) {
	s := testServer(t)
	for i := 0; i < 20; i++ {
		expect(t, request(t, s, "POST", "/api/auth/login", "", map[string]string{"username": "invalid user", "password": "x"}), 401)
	}
	expect(t, request(t, s, "POST", "/api/auth/login", "", map[string]string{"username": "invalid user", "password": "x"}), 429)
}
