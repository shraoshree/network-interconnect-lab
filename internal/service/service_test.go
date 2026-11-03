package service

import (
	"github.com/shraoshree/network-interconnect-lab/internal/checks"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetrics(t *testing.T) {
	s := &Store{}
	s.Set([]checks.Result{{Name: "edge-1", Kind: "tcp", Success: true}})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	s.Handler(nil).ServeHTTP(w, req)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `network_probe_success{name="edge-1",kind="tcp"} 1`) {
		t.Fatalf("unexpected response %d: %s", w.Code, w.Body.String())
	}
}
func TestHealth(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/healthz", nil)
	(&Store{}).Handler(nil).ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}
