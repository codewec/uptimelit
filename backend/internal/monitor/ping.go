package monitor

import (
	"context"
	"net"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

// PingChecker uses unprivileged ICMP echo sockets, without shelling out to ping.
type PingChecker struct{ address string }

func NewPingChecker(address string) *PingChecker { return &PingChecker{address: address} }

func (c *PingChecker) Check(ctx context.Context) (*int, int, bool, string) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	start := time.Now()
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, c.address)
	if err != nil {
		return nil, int(time.Since(start).Milliseconds()), false, connectionErrorCode(err)
	}
	if len(addresses) == 0 {
		return nil, int(time.Since(start).Milliseconds()), false, "dns_error"
	}
	address := addresses[0]
	// Prefer IPv4 for dual-stack hostnames, while allowing explicit IPv6 targets.
	for _, candidate := range addresses {
		if candidate.IP.To4() != nil {
			address = candidate
			break
		}
	}
	pinger := probing.New(c.address)
	pinger.SetIPAddr(&address)
	pinger.SetPrivileged(false)
	pinger.Count = 1
	pinger.Timeout = 10 * time.Second
	if err := pinger.RunWithContext(ctx); err != nil {
		code := connectionErrorCode(err)
		if code == "connection_error" {
			code = "ping_error"
		}
		return nil, int(time.Since(start).Milliseconds()), false, code
	}
	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return nil, int(time.Since(start).Milliseconds()), false, "timeout"
	}
	return nil, int(stats.AvgRtt.Milliseconds()), true, ""
}
