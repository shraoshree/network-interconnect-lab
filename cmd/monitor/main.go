package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shraoshree/network-interconnect-lab/internal/bgp"
	"github.com/shraoshree/network-interconnect-lab/internal/checks"
	"github.com/shraoshree/network-interconnect-lab/internal/service"
)

type Config struct {
	Targets []checks.Target `json:"targets"`
	Routes  []bgp.Route     `json:"routes"`
}

func main() {
	configPath := flag.String("config", "configs/lab.json", "config file path")
	addr := flag.String("listen", ":8080", "HTTP bind address")
	interval := flag.Duration("interval", 15*time.Second, "probe interval")
	timeout := flag.Duration("timeout", 2*time.Second, "per target timeout")
	workers := flag.Int("workers", 4, "maximum concurrent probes")
	flag.Parse()
	if *interval <= 0 || *timeout <= 0 || *workers < 1 {
		log.Fatal("interval, timeout and workers must be positive")
	}
	raw, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		log.Fatal(err)
	}
	for _, t := range cfg.Targets {
		if t.Name == "" || t.Address == "" {
			log.Fatal("target name and address required")
		}
		if t.Kind != "tcp" && t.Kind != "dns" {
			log.Fatal("target kind must be tcp or dns")
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	store := &service.Store{}
	probe := func() {
		r := checks.Run(ctx, checks.NetworkProber{}, cfg.Targets, *workers, *timeout)
		store.Set(r)
		for _, v := range r {
			if !v.Success {
				log.Printf("probe failed name=%q kind=%s error=%q", v.Name, v.Kind, v.Error)
			}
		}
	}
	probe()
	go func() {
		ticker := time.NewTicker(*interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				probe()
			}
		}
	}()
	srv := &http.Server{Addr: *addr, Handler: store.Handler(cfg.Routes), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	log.Printf("monitor listening on %s", *addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
