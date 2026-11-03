// Package checks provides bounded, concurrent TCP and DNS network probes.
package checks

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type Target struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Address string `json:"address"`
}

type Result struct {
	Name      string        `json:"name"`
	Kind      string        `json:"kind"`
	Address   string        `json:"address"`
	Success   bool          `json:"success"`
	Latency   time.Duration `json:"-"`
	Error     string        `json:"error,omitempty"`
	CheckedAt time.Time     `json:"checked_at"`
}

type Prober interface {
	Probe(context.Context, Target) Result
}

type NetworkProber struct {
	Dialer   *net.Dialer
	Resolver *net.Resolver
}

func (p NetworkProber) Probe(ctx context.Context, t Target) Result {
	start := time.Now()
	r := Result{Name: t.Name, Kind: t.Kind, Address: t.Address, CheckedAt: start.UTC()}
	switch strings.ToLower(t.Kind) {
	case "tcp":
		d := p.Dialer
		if d == nil {
			d = &net.Dialer{}
		}
		conn, err := d.DialContext(ctx, "tcp", t.Address)
		if err == nil {
			_ = conn.Close()
		}
		r.Success = err == nil
		if err != nil {
			r.Error = err.Error()
		}
	case "dns":
		resolver := p.Resolver
		if resolver == nil {
			resolver = net.DefaultResolver
		}
		ips, err := resolver.LookupHost(ctx, t.Address)
		r.Success = err == nil && len(ips) > 0
		if err != nil {
			r.Error = err.Error()
		} else if len(ips) == 0 {
			r.Error = "no addresses returned"
		}
	default:
		r.Error = fmt.Sprintf("unsupported check kind %q", t.Kind)
	}
	r.Latency = time.Since(start)
	return r
}

// Run executes each probe with a per-probe timeout, bounded concurrency, and stable result order.
func Run(ctx context.Context, p Prober, targets []Target, workers int, timeout time.Duration) []Result {
	if workers < 1 {
		workers = 1
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	results := make([]Result, len(targets))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				c, cancel := context.WithTimeout(ctx, timeout)
				results[j] = p.Probe(c, targets[j])
				cancel()
			}
		}()
	}
	for i := range targets {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return results
		case jobs <- i:
		}
	}
	close(jobs)
	wg.Wait()
	return results
}
