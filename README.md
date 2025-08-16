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
balancer:
routing_strategy: "simple"

services:
- name: "api"
  path_prefix: "/api/"
  strategy: "round_robin"
  backends:
    - "http://localhost:9001"
    - "http://localhost:9002"
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