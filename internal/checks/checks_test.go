package checks

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestTCPProbe(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	r := NetworkProber{}.Probe(context.Background(), Target{Name: "local", Kind: "tcp", Address: l.Addr().String()})
	if !r.Success {
		t.Fatalf("probe error: %s", r.Error)
	}
}
func TestDNSProbeInvalidName(t *testing.T) {
	r := NetworkProber{}.Probe(context.Background(), Target{Name: "invalid", Kind: "unknown", Address: "host"})
	if r.Success || r.Error == "" {
		t.Fatal("expected failure")
	}
}

type fakeProbe struct{}

func (fakeProbe) Probe(_ context.Context, t Target) Result {
	return Result{Name: t.Name, Success: true}
}
func TestRunStableOrder(t *testing.T) {
	targets := []Target{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	r := Run(context.Background(), fakeProbe{}, targets, 2, time.Second)
	for i, v := range r {
		if v.Name != targets[i].Name {
			t.Fatalf("wrong order: %+v", r)
		}
	}
}
