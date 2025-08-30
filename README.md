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

## Running benchmarks
To run the environment:
```bash
cd benchmark
mkdir results
docker compose up
```
In a new terminal:
[Using wrk]
> You may need to install `wrk` locally
```bash
chmod +x benchmark1.sh
./benchmark1.sh
```
To use k6,
```bash
chmod +x benchmark_k6.sh
./benchmark_k6.sh
```
> You can optimize nginx by changing the files in `backend/nginx` folder and restarting the docker compose setup


## Results
### Stormgate vs NGINX — Out-of-the-Box Round Robin Benchmark

| Metric                   | NGINX (8081)      | Stormgate (8082) | 
|--------------------------|------------------|------------------|
| **Requests/sec**         | 15,926.82         | **56,658.38**     |
| **Total Requests**       | 478,255           | **1,700,332**     |
| **Avg Latency**          | 114.11 ms         | **7.11 ms**       |
| **Max Latency**          | 1.23 s            | **126.11 ms**     |
| **Std Dev (Latency)**    | 200.22 ms         | **3.29 ms**       |
| **Transfer/sec**         | 6.49 MB/s         | **22.42 MB/s**    |
| **Socket Errors (Read)** | **5,689**         | 0                |

✅ **Benchmark Duration:** 30 seconds  
🧪 **Load Config:** 12 threads, 400 connections

### 🌩️ Stormgate vs NGINX — Optimized Round Robin Benchmark

| Metric                   | NGINX (8081)     | Stormgate (8082)  |
|--------------------------|------------------|--------------------|
| **Requests/sec**         | 35,021.07         | 56,147.82           |
| **Total Requests**       | 1,051,091         | 1,684,952           |
| **Avg Latency**          | 11.29 ms          | 7.15 ms             |
| **Max Latency**          | 40.70 ms          | 64.07 ms            |
| **Std Dev (Latency)**    | 2.49 ms           | 3.07 ms             |
| **Transfer/sec**         | 15.30 MB/s        | 22.22 MB/s          |

✅ **Benchmark Duration:** 30 seconds  
🧪 **Load Config:** 12 threads, 400 connections


### 🌩️ Stormgate vs NGINX — Optimized Benchmark (Keepalive: 1000)

| Metric                   | NGINX (8081)     | Stormgate (8082)  |
|--------------------------|------------------|--------------------|
| **Requests/sec**         | 34,396.52         | 55,536.16           |
| **Total Requests**       | 1,032,377         | 1,671,802           |
| **Avg Latency**          | 11.49 ms          | 7.20 ms             |
| **Max Latency**          | 29.26 ms          | 52.37 ms            |
| **Std Dev (Latency)**    | 2.70 ms           | 3.05 ms             |
| **Transfer/sec**         | 15.02 MB/s        | 21.98 MB/s          |

✅ **Benchmark Duration:** 30 seconds  
🧪 **Load Config:** 12 threads, 400 connections  
🔧 **Nginx Keepalive Connections:** 1000


# 🌩️ Stormgate vs NGINX Benchmark Report (k6)

**📅 Date:** Sat Aug 23 19:40:41 BST 2025  
**🧪 Tool:** [k6](https://k6.io/)  
**🖥️ Test Type:** High-throughput load test using `constant-arrival-rate`

---

## 🔧 Test Configuration

| Parameter              | Value            |
|------------------------|------------------|
| Duration               | 30 seconds       |
| Request Rate           | 20,000 req/sec   |
| Executor               | constant-arrival-rate |
| Time Unit              | 1s               |
| Preallocated VUs       | 1000             |
| Max VUs                | 2000 (NGINX), 1000 (Stormgate) |
| Endpoint               | `/round-robin/`  |
| Keep-Alive             | Enabled          |

---

## 📊 Results Summary

| Metric                  | NGINX                            | Stormgate                        |
|-------------------------|----------------------------------|----------------------------------|
| **Total Requests**      | 445,265                          | **600,001**                      |
| **Requests/sec**        | 14,837.78                        | **19,998.95**                    |
| **Avg Latency**         | 92.91 ms                         | **0.82 ms**                      |
| **Median Latency**      | 147.74 ms                        | **0.66 ms**                      |
| **95th Percentile**     | 214.04 ms                        | **1.66 ms**                      |
| **Max Latency**         | 301.64 ms                        | **36.79 ms**                     |
| **Dropped Iterations**  | 154,735                          | **0**                            |
| **HTTP Errors**         | 0                                | 0                                |
| **Data Received**       | 219 MB                           | **265 MB**                       |
| **Data Sent**           | 47 MB                            | **64 MB**                        |

---
If you like this project, please star the repo. Thanks!