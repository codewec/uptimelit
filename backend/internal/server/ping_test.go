package server

import (
	"testing"

	"github.com/codewec/uptimelit/backend/internal/model"
)

func TestValidatePingTarget(t *testing.T) {
	for _, host := range []string{"localhost", "example.com", "example.com.", "my-server.example", "127.0.0.1", "::1", "2001:db8::1", "fe80::1%eth0"} {
		ep := model.Endpoint{Name: "Ping", Type: "ping", URL: host, IntervalSeconds: 10}
		if err := validateEndpoint(&ep); err != nil {
			t.Errorf("valid host %q rejected: %v", host, err)
		}
	}
	for _, host := range []string{"", "https://example.com", "example.com:443", "[::1]:443", "[::1]", "example.com/path", "example com", "-example.com", "example-.com", "a..b", "--help"} {
		ep := model.Endpoint{Name: "Ping", Type: "ping", URL: host, IntervalSeconds: 10}
		if err := validateEndpoint(&ep); err == nil {
			t.Errorf("invalid host %q accepted", host)
		}
	}
}
