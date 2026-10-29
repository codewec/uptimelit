package monitor

import (
	"context"
	"errors"
	"fmt"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestConnectionErrorCode(t *testing.T) {
	for _, tt := range []struct {
		err  error
		want string
	}{
		{nil, ""},
		{context.DeadlineExceeded, "timeout"},
		{&net.DNSError{IsTimeout: true}, "timeout"},
		{&net.DNSError{IsNotFound: true}, "dns_error"},
		{syscall.ECONNREFUSED, "connection_refused"},
		{syscall.ENETUNREACH, "network_unreachable"},
		{syscall.EHOSTUNREACH, "network_unreachable"},
		{syscall.ECONNRESET, "connection_reset"},
		{errors.New("other dial failure"), "connection_error"},
	} {
		err := tt.err
		if err != nil {
			err = fmt.Errorf("dial: %w", &net.OpError{Op: "dial", Net: "tcp", Err: err})
		}
		if got := connectionErrorCode(err); got != tt.want {
			t.Errorf("error %v: got %q, want %q", err, got, tt.want)
		}
	}
}

func TestTCPCheckConnectionAndRefusal(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	defer listener.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	code, _, up, reason := NewTCPChecker(address).Check(ctx)
	if code != nil || !up || reason != "" {
		t.Fatalf("connected: code=%v, up=%v, reason=%q", code, up, reason)
	}
	listener.Close()
	code, _, up, reason = NewTCPChecker(address).Check(ctx)
	if code != nil || up || reason != "connection_refused" {
		t.Fatalf("refused: code=%v, up=%v, reason=%q", code, up, reason)
	}
}
