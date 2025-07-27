# Stormgate [WORK - IN - PROGRESS]

A lightweight L7 Loadbalancer written in Go.
Can Handle storms of traffic. Reasonably well.

## 📦 Features

- ✅ Simple & Hybrid routing strategies
- ⚖️ Load balancers:
    - Round Robin
    - Weighted Round Robin
    - Random
    - IP Hashing (planned)
    - Sticky IP (planned)
- 🔀 Prefix-based routing (fast-match via hybrid bucketing)
- 🔧 YAML-based configuration
- 🚀 Designed for extensibility (GRPC & middleware planned)

---

## 🚀 Getting Started

### 1. Build

```bash
go build ./...
```

### 2. Run
```
go run ./cmd/main.go
```

### 3. Sample Config

Make sure your Config.yaml is present at the project root. Example:

```
server:
  bind_ip: "0.0.0.0"
  bind_port: 10000
  read_timeout_ms: 5000
  write_timeout_ms: 5000

Balancer:
  routing_strategy: 'simple' # or 'hybrid'

services:
  - name: api
    path_prefix: "/api/"
    strategy: random
    backends:
      - http://localhost:9001
      - http://localhost:9002

  - name: auth
    path_prefix: "/auth/"
    strategy: weighted_round_robin
    strategy_config:
      weights: [3, 1]
    backends:
      - http://localhost:9011
      - http://localhost:9012
  
  - name: authv2
    path_prefix: "/auth1/v2/"
    strategy: round_robin
    backends:
      - http://localhost:9011
      - http://localhost:9012
      
  - name: apiv3
    path_prefix: "/api/v3"
    strategy: consistent_hash
    strategy_config:
      source: ip # ip based
      
  - name: apiv4
    path_prefix: "/api/v4"
    strategy: consistent_hash
    strategy_config:
      source: header # header based
      key: "user_id"  # required for header
      fallback_to_ip: true
  
  - name: apiv3
    path_prefix: "/api/v3"
    strategy: consistent_hash
    strategy_config:
      source: cookie # cookie
      key: "user_id"  # optional for cookie. if no key, entire payload from cookie name is taken
      name: "storm_custom" # required for cookie (cookie name)
      fallback_to_ip: true # default is False
      inject_if_missing: true # default is False

  # Add more services here...
```

### 4. Testing

To run all tests:
```
go test ./...
```
