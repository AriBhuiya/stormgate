# 🌩️ Stormgate
*A Lightweight, High-Performance Layer-7 Load Balancer written in Go*

---

### **Why Stormgate?**
Stormgate is a **simple yet powerful** L7 load balancer built for speed, flexibility, and developer-friendliness.  
It supports multiple balancing algorithms, sticky session strategies, health checks, and a simple YAML config — all in one lightweight binary.

---

## ✨ Features
- **Multiple load-balancing strategies**:
    - Round Robin
    - Random
    - Weighted Round Robin
    - Consistent Hash (by IP, Header, or Cookie-Injection)
- **Health checks** (HTTP) with automatic failover
- **Simple routing rules** via path prefixes
- **No external dependencies** — single Go binary

---

## 📦 Quick Start

### 1. Clone & Build
```bash
git clone https://github.com/AriBhuiya/stormgate.git
cd stormgate
go build -o stormgate ./cmd
```

### 2. Create `config.yaml`
Example minimal config:
```yaml
server:
  bind_ip: "0.0.0.0"
  bind_port: 10000
  # Timeouts are in milliseconds
  read_time_out: 5000
  write_time_out: 5000

balancer:
  # "simple" = linear longest-prefix; "hybrid" = hashed buckets + long-prefix list
  routing_strategy: "simple"

services:
  # ---------------------------------------
  # 1) Round Robin
  # ---------------------------------------
  - name: "api-rr"
    path_prefix: "/api/"
    strategy: "round_robin"
    backends:
      - "http://localhost:9001"
      - "http://localhost:9002"
    health:
      # concatenated as <backend>/<health-endpoint>
      health-endpoint: "health"
      type: "http"
      # milliseconds between checks
      frequency: 2000

  # ---------------------------------------
  # 2) Random
  # ---------------------------------------
  - name: "api-random"
    path_prefix: "/random/"
    strategy: "random"
    backends:
      - "http://localhost:9001"
      - "http://localhost:9002"
  #    health:
  #      health-endpoint: "health"
  #      type: "http"
  #      frequency: 2000

  # ---------------------------------------
  # 3) Weighted Round Robin
  #    (weights must align 1:1 with backends)
  # ---------------------------------------
  - name: "api-wrr"
    path_prefix: "/weighted/"
    strategy: "weighted_round_robin"
    backends:
      - "http://localhost:9001"
      - "http://localhost:9002"
    strategy_config:
      weights: [ 3, 1 ]   # 75% to :9001, 25% to :9002
  #    health:
  #      health-endpoint: "health"
  #      type: "http"
  #      frequency: 2000

  # ---------------------------------------
  # 4) Consistent Hash — source: ip
  # ---------------------------------------
  - name: "api-ch-ip"
    path_prefix: "/ch/ip/"
    strategy: "consistent_hash"
    backends:
      - "http://localhost:9001"
      - "http://localhost:9002"
    strategy_config:
      source: "ip"
      # If true, when the primary source is empty it falls back to IP hashing.
      # (For ip source this is a no-op, but included for completeness/compat)
      fallback_to_ip: true
  #    health:
  #      health-endpoint: "health"
  #      type: "http"
  #      frequency: 2000

  # ---------------------------------------
  # 5) Consistent Hash — source: header
  #     Requires a "key" header name.
  # ---------------------------------------
  - name: "api-ch-header"
    path_prefix: "/ch/header/"
    strategy: "consistent_hash"
    backends:
      - "http://localhost:9001"
      - "http://localhost:9002"
    strategy_config:
      source: "header"
      key: "X-User-ID"     # hashed from request header value
      fallback_to_ip: true # when header missing, fall back to IP
  #    health:
  #      health-endpoint: "health"
  #      type: "http"
  #      frequency: 2000

  # ---------------------------------------
  # 6) Consistent Hash — source: cookie
  #     - name: cookie name (defaults to "stormgate-id" in if omitted)
  #     - key:  optional key inside JSON cookie; if cookie is plain string, leave empty
  #     - inject_if_missing: if true, proxy will set a cookie when it's not present
  # ---------------------------------------
  - name: "api-ch-cookie"
    path_prefix: "/ch/cookie/"
    strategy: "consistent_hash"
    backends:
      - "http://localhost:9001"
      - "http://localhost:9002"
    strategy_config:
      source: "cookie"
      name: "stormgate-id"     # optional; defaults to this if not provided
      key: ""                  # optional; for JSON cookie payloads
      inject_if_missing: true  # set a sticky cookie if missing
      fallback_to_ip: false    # if cookie missing and not injecting, whether to fall back to IP
  #    health:
  #      health-endpoint: "health"
  #      type: "http"
  #      frequency: 2000

  # ---------------------------------------
  # 7) Catch‑all (Root) — Round Robin
  # ---------------------------------------
  - name: "root"
    path_prefix: "/"
    strategy: "round_robin"
    backends:
      - "http://localhost:9001"
      - "http://localhost:9002"
  #    health:
  #      health-endpoint: "health"
  #      type: "http"
  #      frequency: 2000
```

### 3. Run
```
./stormgate
```
Stormgate will listen on `0.0.0.0:10000` and forward requests according to `config.yaml`.

---

## ⚙️ Configuration Guide

A full example config showing **all features** is in `sample_config.yaml`.

## Docker
```bash
docker compose up
```


---

## 🧪 Testing

Stormgate ships with **self-contained smoke tests** in `tests/`:
- `smoke_core.sh` — verifies all balancing strategies
- `smoke_health.sh` — verifies health checks and failover

Run from the repo root:
```bash
bash tests/smoke_core.sh
bash tests/smoke_health.sh
```

These scripts:
- Spin up mock backends on `:9001` / `:9002`
- Build and run Stormgate with a temporary config
- Verify routing correctness
- Clean up everything on exit

---

## 📊 Example Output
```bash
>> starting stormgate
>> test: round robin /api/
RR counts: 10 vs 10
>> test: random /random/
Random counts: 9 vs 11
✅ ALL SMOKE TESTS PASSED
```

---

## 🛠 Development

### Run Unit Tests
```bash
go test ./...
```