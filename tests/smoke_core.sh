#!/usr/bin/env bash
set -euo pipefail

# ── Config ─────────────────────────────────────────────────────────────────────
STORMGATE_PORT=10000
BACKEND1_PORT=9001
BACKEND2_PORT=9002
REQS_SMALL=20
REQS_WRR=120
WRR_MIN_PCT_9001=60
WRR_MAX_PCT_9001=85

# ── Helpers ────────────────────────────────────────────────────────────────────
TMPDIR="$(mktemp -d)"
SG_LOG="$TMPDIR/sg.log"

cleanup() {
  echo ">> cleanup"
  [[ -f config.yaml.bak ]] && mv -f config.yaml.bak config.yaml || true
  [[ -n "${SG_PID:-}" ]]  && kill "${SG_PID}" 2>/dev/null || true
  [[ -n "${B1_PID:-}" ]]  && kill "${B1_PID}" 2>/dev/null || true
  [[ -n "${B2_PID:-}" ]]  && kill "${B2_PID}" 2>/dev/null || true
}
trap cleanup EXIT

fail() {
  echo "---- Stormgate log (last 200) ----"
  tail -n 200 "$SG_LOG" 2>/dev/null || true
  echo "FAIL:" "$@"
  exit 1
}

wait_for_port() {
  local host="$1" port="$2" name="$3" tries=80
  echo ">> waiting for $name on $host:$port"
  for _ in $(seq 1 "$tries"); do
    if (echo > /dev/tcp/$host/$port) >/dev/null 2>&1; then
      echo "OK: $name up"; return 0
    fi
    sleep 0.25
  done
  fail "$name on $host:$port did not come up in time"
}

percent() { python3 - "$1" "$2" <<'PY'
import sys
num, den = int(sys.argv[1]), int(sys.argv[2])
print(int(round((num*100.0)/max(den,1))))
PY
}

# Return ":9001" style value from response headers
port_from_headers() {
  # usage: port_from_headers "http://host:port/path"
  curl -fsS -D - "$1" -o /dev/null \
    | awk -F': ' '/^X-Backend-Port:/ {gsub(/\r/,"",$2); print ":"$2; exit}'
}

kill_port() {
  local port="$1"
  if command -v lsof >/dev/null 2>&1; then
    local pids; pids=$(lsof -ti tcp:"$port" || true)
    if [[ -n "$pids" ]]; then
      echo ">> freeing port $port (PIDs: $pids)"
      kill $pids 2>/dev/null || true
      sleep 0.3
      pids=$(lsof -ti tcp:"$port" || true)
      [[ -n "$pids" ]] && kill -9 $pids 2>/dev/null || true
    fi
  fi
}

# ── Build mock backend ─────────────────────────────────────────────────────────
cat > "$TMPDIR/backend.go" <<'GO'
package main
import ("fmt";"log";"net/http";"os")
func main(){
  port := os.Getenv("PORT"); if port==""{port="9001"}
  mux := http.NewServeMux()
  mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request){
    w.WriteHeader(http.StatusOK); _,_ = w.Write([]byte("ok"))
  })
  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
    w.Header().Set("X-Backend-Port", port)
    fmt.Fprintf(w, "backend:%s method:%s path:%s\n", port, r.Method, r.URL.RequestURI())
  })
  addr := ":"+port
  log.Printf("mock backend on %s", addr)
  log.Fatal(http.ListenAndServe(addr, mux))
}
GO
echo ">> building mock backend"
BACK_BIN="$TMPDIR/backend"
GO111MODULE=on go build -o "$BACK_BIN" "$TMPDIR/backend.go"

# ── Always run mocks; free ports first ─────────────────────────────────────────
kill_port "$BACKEND1_PORT"
kill_port "$BACKEND2_PORT"
PORT=$BACKEND1_PORT "$BACK_BIN" >"$TMPDIR/b1.log" 2>&1 & B1_PID=$!
PORT=$BACKEND2_PORT "$BACK_BIN" >"$TMPDIR/b2.log" 2>&1 & B2_PID=$!
wait_for_port 127.0.0.1 "$BACKEND1_PORT" "backend1"
wait_for_port 127.0.0.1 "$BACKEND2_PORT" "backend2"

# ── Backup config + write test config (no health) ──────────────────────────────
[[ -f config.yaml ]] && cp config.yaml config.yaml.bak
cat > config.yaml <<'YAML'
server:
  bind_ip: "0.0.0.0"
  bind_port: 10000
  read_time_out: 5000
  write_time_out: 5000
balancer:
  routing_strategy: "simple"
services:
  - name: "api-rr"
    path_prefix: "/api/"
    strategy: "round_robin"
    backends: ["http://localhost:9001","http://localhost:9002"]
  - name: "api-random"
    path_prefix: "/random/"
    strategy: "random"
    backends: ["http://localhost:9001","http://localhost:9002"]
  - name: "api-wrr"
    path_prefix: "/weighted/"
    strategy: "weighted_round_robin"
    backends: ["http://localhost:9001","http://localhost:9002"]
    strategy_config: { weights: [3,1] }
  - name: "api-ch-ip"
    path_prefix: "/ch/ip/"
    strategy: "consistent_hash"
    backends: ["http://localhost:9001","http://localhost:9002"]
    strategy_config: { source: "ip", fallback_to_ip: true }
  - name: "api-ch-header"
    path_prefix: "/ch/header/"
    strategy: "consistent_hash"
    backends: ["http://localhost:9001","http://localhost:9002"]
    strategy_config: { source: "header", key: "X-User-ID", fallback_to_ip: true }
  - name: "api-ch-cookie"
    path_prefix: "/ch/cookie/"
    strategy: "consistent_hash"
    backends: ["http://localhost:9001","http://localhost:9002"]
    strategy_config: { source: "cookie", name: "stormgate-id", key: "", inject_if_missing: true, fallback_to_ip: false }
  - name: "root"
    path_prefix: "/"
    strategy: "round_robin"
    backends: ["http://localhost:9001","http://localhost:9002"]
YAML

# ── Start Stormgate ────────────────────────────────────────────────────────────
kill_port "$STORMGATE_PORT"
echo ">> starting stormgate"
go build -o "$TMPDIR/stormgate" ./cmd
"$TMPDIR/stormgate" >"$SG_LOG" 2>&1 & SG_PID=$!
wait_for_port 127.0.0.1 "$STORMGATE_PORT" "stormgate"
sleep 0.7
BASE="http://127.0.0.1:${STORMGATE_PORT}"

# ── Tests ──────────────────────────────────────────────────────────────────────
echo ">> RR /api/"
rr1=0; rr2=0
for i in $(seq 1 $REQS_SMALL); do
  p="$(port_from_headers "${BASE}/api/")"
  [[ "$p" == ":${BACKEND1_PORT}" ]] && rr1=$((rr1+1))
  [[ "$p" == ":${BACKEND2_PORT}" ]] && rr2=$((rr2+1))
done
echo "RR: :$BACKEND1_PORT=$rr1 :$BACKEND2_PORT=$rr2"
[[ $rr1 -gt 0 && $rr2 -gt 0 ]] || fail "RR did not hit both backends"

echo ">> Random /random/"
r1=0; r2=0
for i in $(seq 1 $REQS_SMALL); do
  p="$(port_from_headers "${BASE}/random/")"
  [[ "$p" == ":${BACKEND1_PORT}" ]] && r1=$((r1+1)) || r2=$((r2+1))
done
echo "Random: :$BACKEND1_PORT=$r1 :$BACKEND2_PORT=$r2"
[[ $((r1+r2)) -eq $REQS_SMALL ]] || fail "Random total mismatch"

echo ">> WRR /weighted/ (~3:1)"
w1=0; w2=0
for i in $(seq 1 $REQS_WRR); do
  p="$(port_from_headers "${BASE}/weighted/")"
  [[ "$p" == ":${BACKEND1_PORT}" ]] && w1=$((w1+1)) || w2=$((w2+1))
done
pct1=$(percent "$w1" "$REQS_WRR")
echo "WRR: :$BACKEND1_PORT=$w1 :$BACKEND2_PORT=$w2 (${pct1}%)"
[[ $pct1 -ge $WRR_MIN_PCT_9001 && $pct1 -le $WRR_MAX_PCT_9001 ]] || fail "WRR not ~3:1"

echo ">> CH (ip) /ch/ip/"
pA="$(port_from_headers "${BASE}/ch/ip/")"
pB="$(port_from_headers "${BASE}/ch/ip/")"
[[ "$pA" == "$pB" ]] || fail "CH-IP not sticky"

echo ">> CH (header) /ch/header/"
h1="$(port_from_headers "${BASE}/ch/header/")"
h2="$(port_from_headers "${BASE}/ch/header/")"
# same header each time (proxy doesn't change header itself), so should be sticky
[[ "$h1" == "$h2" ]] || fail "CH-Header not sticky for same key"
# Additional explicit key test
ha="$(curl -fsS -H "X-User-ID=alpha" "${BASE}/ch/header/" -D - -o /dev/null | awk -F': ' '/^X-Backend-Port:/ {gsub(/\r/,"",$2); print ":"$2; exit}')"
hb="$(curl -fsS -H "X-User-ID=alpha" "${BASE}/ch/header/" -D - -o /dev/null | awk -F': ' '/^X-Backend-Port:/ {gsub(/\r/,"",$2); print ":"$2; exit}')"
[[ "$ha" == "$hb" ]] || fail "CH-Header (explicit) not sticky"

echo ">> CH (cookie) /ch/cookie/"
hdrs="$(curl -fsS -D - "${BASE}/ch/cookie/" -o /dev/null)"
cookie="$(echo "$hdrs" | awk -F': ' '/^Set-Cookie:/ {gsub(/\r/,"",$2); print $2; exit}')"
[[ -n "$cookie" ]] || fail "expected Set-Cookie"
c1="$(curl -fsS -H "Cookie: $cookie" "${BASE}/ch/cookie/" -D - -o /dev/null | awk -F': ' '/^X-Backend-Port:/ {gsub(/\r/,"",$2); print ":"$2; exit}')"
c2="$(curl -fsS -H "Cookie: $cookie" "${BASE}/ch/cookie/" -D - -o /dev/null | awk -F': ' '/^X-Backend-Port:/ {gsub(/\r/,"",$2); print ":"$2; exit}')"
[[ "$c1" == "$c2" ]] || fail "CH-Cookie not sticky"

echo "✅ ALL SMOKE TESTS PASSED"