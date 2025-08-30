#!/bin/bash

# Output file for CPU usage
CPU_LOG="../results/cpu_usage_nginx.log"
BENCH_LOG="../results/k6_benchmark_nginx.log"

# Clear old logs
> "$CPU_LOG"
> "$BENCH_LOG"

# Monitor CPU every 1s in the background
echo "📈 Monitoring CPU usage..."
top -l 0 -s 1 -n 30 -stats cpu > "$CPU_LOG" &   # macOS

CPU_MONITOR_PID=$!

# Run k6 test
echo "🚀 Running k6 test..."
TARGET_URL=http://localhost:8081/round-robin/ k6 run k6test.js > "$BENCH_LOG"

# Stop CPU monitoring
kill $CPU_MONITOR_PID

echo "✅ Benchmark and CPU usage recorded."
echo "📝 CPU: $CPU_LOG"
echo "📝 K6:  $BENCH_LOG"