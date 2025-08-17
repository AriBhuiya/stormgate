#!/bin/bash

OUTPUT_FILE="../results/k6_results.txt"
TEST_FILE="./k6test.js"

mkdir -p ../results

echo "🌩️ Stormgate vs NGINX Benchmark (k6) - $(date)" > "$OUTPUT_FILE"
echo "-------------------------------------------------------------" >> "$OUTPUT_FILE"

# NGINX Test
echo "🔁 Testing NGINX on http://localhost:8081/round-robin/..." | tee -a "$OUTPUT_FILE"
k6 run -e TARGET_URL=http://localhost:8081/round-robin/ "$TEST_FILE" >> "$OUTPUT_FILE"

echo -e "\n-------------------------------------------------------------\n" >> "$OUTPUT_FILE"

# Stormgate Test
echo "⚡ Testing Stormgate on http://localhost:8082/round-robin/..." | tee -a "$OUTPUT_FILE"
k6 run -e TARGET_URL=http://localhost:8082/round-robin/ "$TEST_FILE" >> "$OUTPUT_FILE"

echo -e "\n✅ k6 benchmark complete. Results saved to $OUTPUT_FILE"