#!/bin/bash

# Config
DURATION="30s"
THREADS=12
CONNECTIONS=400
OUTPUT_FILE="../results/benchmark_results_out_of_box.txt"
NGINX_URL="http://localhost:8081/round-robin/"
STORMGATE_URL="http://localhost:8082/round-robin/"

# Clear previous results
echo "🌩️ Stormgate vs NGINX Benchmark - $(date)" > "$OUTPUT_FILE"
echo "Load Test Config: Duration=$DURATION, Threads=$THREADS, Connections=$CONNECTIONS" >> "$OUTPUT_FILE"
echo "-------------------------------------------------------------" >> "$OUTPUT_FILE"

# Run NGINX test
echo "🔁 Testing NGINX on $NGINX_URL..." | tee -a "$OUTPUT_FILE"
echo -e "\n[NGINX Results]" >> "$OUTPUT_FILE"
wrk -t$THREADS -c$CONNECTIONS -d$DURATION "$NGINX_URL" >> "$OUTPUT_FILE"

# Spacer
echo -e "\n-------------------------------------------------------------\n" >> "$OUTPUT_FILE"

# Run Stormgate test
echo "⚡ Testing Stormgate on $STORMGATE_URL..." | tee -a "$OUTPUT_FILE"
echo -e "\n[Stormgate Results]" >> "$OUTPUT_FILE"
wrk -t$THREADS -c$CONNECTIONS -d$DURATION "$STORMGATE_URL" >> "$OUTPUT_FILE"

# Done
echo -e "\n✅ Benchmark complete. Results saved to $OUTPUT_FILE\n" | tee -a "$OUTPUT_FILE"