package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/shraoshree/network-interconnect-lab/internal/bgp"
	"github.com/shraoshree/network-interconnect-lab/internal/checks"
)

type Snapshot struct {
	Results   []checks.Result `json:"results"`
	UpdatedAt time.Time       `json:"updated_at"`
}
type Store struct {
	mu      sync.RWMutex
	current Snapshot
}

func (s *Store) Set(r []checks.Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = Snapshot{Results: append([]checks.Result(nil), r...), UpdatedAt: time.Now().UTC()}
}
func (s *Store) Get() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := s.current
	v.Results = append([]checks.Result(nil), v.Results...)
	return v
}
func (s *Store) Handler(routes []bgp.Route) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("GET /api/v1/checks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s.Get())
	})
	mux.HandleFunc("GET /api/v1/routes/best", func(w http.ResponseWriter, r *http.Request) {
		best, err := bgp.Best(routes)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		_ = json.NewEncoder(w).Encode(best)
	})
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		snap := s.Get()
		_, _ = fmt.Fprintln(w, "# HELP network_probe_success Whether the latest network probe succeeded (1) or failed (0).")
		_, _ = fmt.Fprintln(w, "# TYPE network_probe_success gauge")
		_, _ = fmt.Fprintln(w, "# HELP network_probe_duration_seconds Duration of the latest network probe in seconds.")
		_, _ = fmt.Fprintln(w, "# TYPE network_probe_duration_seconds gauge")
		for _, v := range snap.Results {
			name := metricEscape(v.Name)
			kind := metricEscape(v.Kind)
			value := 0
			if v.Success {
				value = 1
			}
			_, _ = fmt.Fprintf(w, "network_probe_success{name=\"%s\",kind=\"%s\"} %d\n", name, kind, value)
			_, _ = fmt.Fprintf(w, "network_probe_duration_seconds{name=\"%s\",kind=\"%s\"} %.6f\n", name, kind, v.Latency.Seconds())
		}
	})
	return mux
}
func metricEscape(s string) string {
	out := ""
	for _, r := range s {
		switch r {
		case '\\':
			out += "\\\\"
		case '"':
			out += "\\\""
		case '\n':
			out += "\\n"
		default:
			out += string(r)
		}
	}
	return out
}
