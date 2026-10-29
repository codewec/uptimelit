package server

import (
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestFrontendRoutes(t *testing.T) {
	s := testServer(t)
	s.frontend = fstest.MapFS{
		"index.html":          {Data: []byte(`<!doctype html><div id="app"></div>`)},
		"assets/app-ab123.js": {Data: []byte(`console.log("uptimelit")`)},
		".gitkeep":            {Data: nil},
	}
	for _, route := range []string{"/", "/login", "/profile", "/admin", "/notifiers"} {
		w := request(t, s, "GET", route, "", nil)
		expect(t, w, 200)
		if !strings.Contains(w.Body.String(), `id="app"`) || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("invalid SPA response for %s", route)
		}
	}
	w := request(t, s, "GET", "/assets/app-ab123.js", "", nil)
	expect(t, w, 200)
	if !strings.Contains(w.Header().Get("Cache-Control"), "immutable") {
		t.Fatal("hashed asset not cached")
	}
	for _, route := range []string{"/assets/missing.js", "/favicon-missing.png", "/.gitkeep", "/api/missing"} {
		w = request(t, s, "GET", route, "", nil)
		expect(t, w, 404)
		if strings.Contains(w.Body.String(), `id="app"`) {
			t.Fatalf("SPA returned for %s", route)
		}
	}
	expect(t, request(t, s, "GET", "/api/endpoints", "", nil), 401)
}
func TestFrontendNotBuilt(t *testing.T) {
	s := testServer(t)
	s.frontend = fstest.MapFS{".gitkeep": {Data: nil}}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	expect(t, w, 503)
}
