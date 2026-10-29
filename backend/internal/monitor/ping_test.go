package monitor

import (
	"context"
	"testing"
	"time"
)

func TestPingLoopback(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	code, elapsed, up, reason := NewPingChecker("127.0.0.1").Check(ctx)
	if code != nil || !up || reason != "" || elapsed < 0 {
		t.Fatalf("loopback ping: code=%v time=%d up=%v reason=%q", code, elapsed, up, reason)
	}
}

func TestPingExpiredContext(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	start := time.Now()
	_, _, up, reason := NewPingChecker("127.0.0.1").Check(ctx)
	if up || reason != "timeout" || time.Since(start) > time.Second {
		t.Fatalf("expired ping: up=%v reason=%q elapsed=%s", up, reason, time.Since(start))
	}
}
