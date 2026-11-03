# Network Interconnect Lab — Go Monitoring & BGP Route Selection

A portfolio project for systems/network-infrastructure software engineering. It combines a concurrent Go TCP/DNS connectivity monitor, Prometheus-compatible metrics, a provisioned Grafana dashboard, containerized deployment, and a **simplified offline BGP best-path selection exercise**.

> **Scope:** This project does NOT implement BGP sessions, router configuration, VLAN provisioning, GRE/VXLAN encapsulation, NetBox/IPAM integrations, or physical cross-connect orchestration. The route selector is a learning simulation, not production routing software.

## Architecture

```text
config targets ─→ bounded Go workers ─→ TCP dial / DNS lookup
                              │
                              ├─→ GET /api/v1/checks (JSON)
                              └─→ GET /metrics ─→ Prometheus ─→ Grafana
sample BGP candidates ─→ route selection algorithm ─→ GET /api/v1/routes/best
```

## Highlights

- Go worker pool with contexts, per-target deadlines, concurrent probes, deterministic results, and clean shutdown.
- Real TCP reachability checks and DNS resolution; no raw sockets or elevated privileges required.
- HTTP health endpoint, JSON results API, and scrape-friendly Prometheus gauge metrics.
- Docker multi-stage build, Docker Compose stack, provisioned Grafana dashboard, GitHub Actions Go tests and vet.
- Offline BGP-route selection **simulation** prioritizing LOCAL_PREF, AS_PATH length, and MED, with deterministic tie-breaking. This intentionally simplifies real BGP policy and does **not** exchange BGP messages.
- Unit tests for route preferences, validation, health, metrics, TCP reachability, worker ordering.

## Quick start

Requires Go 1.23+.

```bash
go test ./...
go run ./cmd/monitor -config configs/lab.json
```

In another terminal:

```bash
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/api/v1/checks
curl -s http://localhost:8080/api/v1/routes/best
curl -s http://localhost:8080/metrics
```

Or run Grafana and Prometheus together:

```bash
GRAFANA_ADMIN_PASSWORD='choose-a-strong-local-password' docker compose up --build
```

Monitor: http://localhost:8080, Prometheus: http://localhost:9090, Grafana: http://localhost:3000. Default Grafana username: `admin`.

**Note:** The example DNS target requires outbound DNS. Network restrictions can make it fail; the monitor records this as an unsuccessful probe rather than crashing.

## API

| Endpoint | Purpose |
|---|---|
| `GET /healthz` | process health |
| `GET /api/v1/checks` | latest connectivity snapshot |
| `GET /api/v1/routes/best` | demonstration BGP best route |
| `GET /metrics` | Prometheus text metrics |

## Routing example

`configs/lab.json` includes three **documentation-only** BGP candidates for the same prefix (`203.0.113.0/24`), with reserved sample next hops (`192.0.2.0/24`). The selector chooses highest local preference, then shortest AS path, then lowest MED, then lexical next-hop address. Real routers apply more tie-break rules, and MED comparison is usually restricted by neighboring AS. Never deploy this as a real routing decision engine.

## Design choices and next steps

The worker pool caps concurrent socket operations, contexts enforce deadlines, snapshots use a mutex, and Prometheus samples expose outcomes/latencies. A real enterprise implementation would additionally need authenticated configuration, hardened target allowlists, persistent historical results, bounded metrics label cardinality, structured logging, tracing, a genuine BGP integration (such as GoBGP), and integration tests involving network namespaces and virtual devices.

## Resume-ready description

> Built a Go-based network observability lab with concurrent TCP/DNS probes, timeout-aware worker pools, Prometheus metrics, Grafana dashboards, Docker Compose, CI tests, and an offline BGP best-path selection simulator.

This is a personal portfolio project, **not** professional Nokia experience.
