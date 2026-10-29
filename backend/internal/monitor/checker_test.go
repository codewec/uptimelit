package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHTTPChecker(t *testing.T) {
	for _, tt := range []struct {
		name     string
		headCode int
		getCode  int
		wantCode int
		wantUp   bool
		methods  []string
	}{
		{"HEAD succeeds", 200, 500, 200, true, []string{"HEAD"}},
		{"HEAD unsupported GET succeeds", 405, 200, 200, true, []string{"HEAD", "GET"}},
		{"HEAD unsupported GET fails", 405, 503, 503, false, []string{"HEAD", "GET"}},
		{"HEAD server error", 503, 200, 503, false, []string{"HEAD"}},
		{"HEAD forbidden", 403, 200, 403, false, []string{"HEAD"}},
		{"redirect counts as up", 302, 500, 302, true, []string{"HEAD"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var methods []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				methods = append(methods, r.Method)
				if r.Method == http.MethodHead {
					w.WriteHeader(tt.headCode)
					return
				}
				w.WriteHeader(tt.getCode)
			}))
			defer server.Close()

			code, _, up := NewHTTPChecker(server.URL).Check(context.Background())
			if code == nil || *code != tt.wantCode || up != tt.wantUp {
				t.Fatalf("Check() = (%v, up=%v), want code=%d, up=%v", code, up, tt.wantCode, tt.wantUp)
			}
			if !reflect.DeepEqual(methods, tt.methods) {
				t.Fatalf("methods = %v, want %v", methods, tt.methods)
			}
		})
	}
}
