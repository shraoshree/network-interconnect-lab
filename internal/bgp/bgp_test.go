package bgp

import "testing"

func TestBestPrefThenASPath(t *testing.T) {
	routes := []Route{{Prefix: "203.0.113.0/24", NextHop: "192.0.2.1", LocalPref: 100, ASPath: []uint32{1}, MED: 0}, {Prefix: "203.0.113.0/24", NextHop: "192.0.2.2", LocalPref: 200, ASPath: []uint32{1, 2}, MED: 20}, {Prefix: "203.0.113.0/24", NextHop: "192.0.2.3", LocalPref: 200, ASPath: []uint32{1}, MED: 90}}
	got, err := Best(routes)
	if err != nil || got.NextHop != "192.0.2.3" {
		t.Fatalf("got %+v error %v", got, err)
	}
}
func TestBestRejectsMixedPrefixes(t *testing.T) {
	_, err := Best([]Route{{Prefix: "10.0.0.0/8", NextHop: "192.0.2.1"}, {Prefix: "10.1.0.0/16", NextHop: "192.0.2.2"}})
	if err == nil {
		t.Fatal("expected error")
	}
}
func TestBestEmpty(t *testing.T) {
	if _, err := Best(nil); err == nil {
		t.Fatal("expected error")
	}
}
