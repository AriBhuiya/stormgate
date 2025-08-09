#!/usr/bin/env bash
set -euo pipefail

# ── Config ─────────────────────────────────────────────────────────────────────
STORMGATE_PORT=10000
BACKEND1_PORT=9001
BACKEND2_PORT=9002
REQS_SMALL=20
HEALTH_FREQ_MS=1500     # must match config below
HEALTH_WAIT_MULT=2      # waits ~3s for convergence (2 * 1500ms)

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

sleep_health() {
  # wait ~ (HEALTH_FREQ_MS * HEALTH_WAIT_MULT)
  local sec=$(( (HEALTH_FREQ_MS * HEALTH_WAIT_MULT) / 1000 + 1 ))
  sleep "$sec"
}

# Read ":9001" from response headers
port_from_headers() {
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

curl_s(){ curl -fsS "$@"; }

# ── Build mock backend with /health ────────────────────────────────────────────
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

# ── Backup + write config (health enabled for ALL services) ────────────────────
[[ -f config.yaml ]] && cp config.yaml config.yaml.bak
cat > config.yaml <<YAML
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
    health: { health-endpoint: "health", type: "http", frequency: ${HEALTH_FREQ_MS} }

  - name: "api-random"
    path_prefix: "/random/"
    strategy: "random"
    backends: ["http://localhost:9001","http://localhost:9002"]
    health: { health-endpoint: "health", type: "http", frequency: ${HEALTH_FREQ_MS} }

  - name: "api-wrr"
    path_prefix: "/weighted/"
    strategy: "weighted_round_robin"
    backends: ["http://localhost:9001","http://localhost:9002"]
    strategy_config: { weights: [3,1] }
    health: { health-endpoint: "health", type: "http", frequency: ${HEALTH_FREQ_MS} }

  - name: "api-ch-ip"
    path_prefix: "/ch/ip/"
    strategy: "consistent_hash"
    backends: ["http://localhost:9001","http://localhost:9002"]
    strategy_config: { source: "ip", fallback_to_ip: true }
    health: { health-endpoint: "health", type: "http", frequency: ${HEALTH_FREQ_MS} }

  - name: "api-ch-header"
    path_prefix: "/ch/header/"
    strategy: "consistent_hash"
    backends: ["http://localhost:9001","http://localhost:9002"]
    strategy_config: { source: "header", key: "X-User-ID", fallback_to_ip: true }
    health: { health-endpoint: "health", type: "http", frequency: ${HEALTH_FREQ_MS} }

  - name: "api-ch-cookie"
    path_prefix: "/ch/cookie/"
    strategy: "consistent_hash"
    backends: ["http://localhost:9001","http://localhost:9002"]
    strategy_config: { source: "cookie", name: "stormgate-id", key: "", inject_if_missing: true, fallback_to_ip: false }
    health: { health-endpoint: "health", type: "http", frequency: ${HEALTH_FREQ_MS} }

  - name: "root"
    path_prefix: "/"
    strategy: "round_robin"
    backends: ["http://localhost:9001","http://localhost:9002"]
    health: { health-endpoint: "health", type: "http", frequency: ${HEALTH_FREQ_MS} }
YAML

# ── Start Stormgate ────────────────────────────────────────────────────────────
kill_port "$STORMGATE_PORT"
echo ">> starting stormgate"
go build -o "$TMPDIR/stormgate" ./cmd
"$TMPDIR/stormgate" >"$SG_LOG" 2>&1 & SG_PID=$!
wait_for_port 127.0.0.1 "$STORMGATE_PORT" "stormgate"
sleep 1
BASE="http://127.0.0.1:${STORMGATE_PORT}"

# ── Baseline quick sanity (one call per service) ───────────────────────────────
paths=( "/api/" "/random/" "/weighted/" "/ch/ip/" "/ch/header/" "/ch/cookie/" "/" )
for p in "${paths[@]}"; do
  echo ">> baseline: $p"
  curl -fsS "$BASE$p" >/dev/null
done

# ── Kill backend :9002 and wait for health to converge ─────────────────────────
echo ">> stopping backend :$BACKEND2_PORT"
kill "$B2_PID" 2>/dev/null || true
sleep_health

# ── Verify all services avoid :9002 ────────────────────────────────────────────
check_not_9002() { # path, optional curl args (e.g., -H ...)
  local path="$1"; shift || true
  local hit2=0
  for i in $(seq 1 $REQS_SMALL); do
    p="$(curl -fsS "$@" "$BASE$path" -D - -o /dev/null | awk -F': ' '/^X-Backend-Port:/ {gsub(/\r/,"",$2); print ":"$2; exit}')"
    [[ "$p" == ":${BACKEND2_PORT}" ]] && hit2=$((hit2+1))
  done
  [[ $hit2 -eq 0 ]] || fail "$path still hits :${BACKEND2_PORT}"
}

echo ">> verifying each service avoids :$BACKEND2_PORT"
check_not_9002 "/api/"
check_not_9002 "/random/"
check_not_9002 "/weighted/"
check_not_9002 "/"
check_not_9002 "/ch/ip/"
check_not_9002 "/ch/header/" -H "X-User-ID=alpha"
check_not_9002 "/ch/header/" -H "X-User-ID=beta"

# CH-cookie: ensure freshly injected cookies never map to :9002
for i in $(seq 1 $REQS_SMALL); do
  hdrs="$(curl -fsS -D - "$BASE/ch/cookie/" -o /dev/null)"
  cookie="$(echo "$hdrs" | awk -F': ' '/^Set-Cookie:/ {gsub(/\r/,"",$2); print $2; exit}')"
  [[ -n "$cookie" ]] || fail "expected Set-Cookie during CH-cookie check"
  p="$(curl -fsS -H "Cookie: $cookie" "$BASE/ch/cookie/" -D - -o /dev/null | awk -F': ' '/^X-Backend-Port:/ {gsub(/\r/,"",$2); print ":"$2; exit}')"
  [[ "$p" != ":${BACKEND2_PORT}" ]] || fail "CH-cookie still hit :${BACKEND2_PORT}"
done

# ── Restart backend :9002, wait again, verify recovery ─────────────────────────
echo ">> restarting backend :$BACKEND2_PORT"
PORT=$BACKEND2_PORT "$BACK_BIN" >"$TMPDIR/b2.log" 2>&1 & B2_PID=$!
wait_for_port 127.0.0.1 "$BACKEND2_PORT" "backend2"
sleep_health

# Quick recovery check on /api/ (should see :9002 at least once)
a1=0; a2=0
for i in $(seq 1 $REQS_SMALL); do
  p="$(port_from_headers "$BASE/api/")"
  [[ "$p" == ":${BACKEND1_PORT}" ]] && a1=$((a1+1))
  [[ "$p" == ":${BACKEND2_PORT}" ]] && a2=$((a2+1))
done
echo "After restart /api/ hits: :$BACKEND1_PORT=$a1 :$BACKEND2_PORT=$a2"
[[ $a2 -gt 0 ]] || fail "recovered backend did not receive any traffic"

echo "✅ SMOKE HEALTH PASSED"