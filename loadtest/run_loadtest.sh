#!/bin/bash
set -e

# Build binaries
echo "Building binaries..."
go build -o janus-proxy ../cmd/proxy
go build -o dummy-backend ./dummy_backend

# Cleanup previous runs
pkill -f dummy-backend || true
pkill -f janus-proxy || true

# Start dummy backends
echo "Starting backends..."
./dummy-backend -port 8081 &
PID_B1=$!
./dummy-backend -port 8082 &
PID_B2=$!

# Wait for backends
sleep 1

# Create config if not exists
cat <<EOF > config_load.json
{
  "port": 8080,
  "strategy": "round_robin",
  "health_check_time": 1,
  "backends": [
    { "url": "http://localhost:8081", "weight": 1 },
    { "url": "http://localhost:8082", "weight": 1 }
  ]
}
EOF

# Start Janus
echo "Starting Janus..."
./janus-proxy -config config_load.json > janus.log 2>&1 &
PID_PROXY=$!

# Wait for Janus to be ready
echo "Waiting for Janus to start..."
for i in {1..10}; do
    if curl -s http://127.0.0.1:8080 > /dev/null; then
        echo "Janus is up!"
        break
    fi
    sleep 1
done

# Run Load Test
echo "Running Stress Test with Hey (Target: 100k requests, 500 concurrent)..."
# -n 100000 requests, -c 500 concurrency
if ! ~/go/bin/hey -n 100000 -c 500 http://127.0.0.1:8080/ > load_test_results.txt; then
    echo "Hey failed. checking logs:"
    cat janus.log
    exit 1
fi

echo "Load test complete. Results:"
grep "Requests/sec" load_test_results.txt
grep "Latency" load_test_results.txt
grep "Error distribution" -A 5 load_test_results.txt || true

# Cleanup
kill $PID_B1 $PID_B2 $PID_PROXY
rm janus-proxy dummy-backend
rm config_load.json
