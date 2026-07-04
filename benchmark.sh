#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

echo "=========================================================="
echo "🚀 Order Matching Engine - Automated Benchmark Suite"
echo "=========================================================="

# 1. Check and Install 'hey' if missing
export PATH=$PATH:$(go env GOPATH)/bin
if ! command -v hey &> /dev/null; then
    echo "📥 'hey' load testing tool not found. Installing..."
    go install github.com/rakyll/hey@latest
    echo "✅ 'hey' installed successfully."
else
    echo "✅ 'hey' is already installed."
fi

# 2. Prepare Dummy Data
echo "📁 Generating dummy order payloads..."
mkdir -p dummy-data
cat <<EOF > dummy-data/buy.json
{
  "symbol": "AAPL",
  "side": "BUY",
  "type": "LIMIT",
  "price": 15050,
  "quantity": 100
}
EOF
cat <<EOF > dummy-data/sell.json
{
  "symbol": "AAPL",
  "side": "SELL",
  "type": "LIMIT",
  "price": 15050,
  "quantity": 100
}
EOF

# 3. Clean up any existing instances on port 8080
echo "🧹 Checking for existing processes on port 8080..."
lsof -ti:8080 | xargs kill -9 2>/dev/null || true

# 4. Start the Engine in the background
echo "⚡ Compiling and Starting the Matching Engine binary..."
go build -o engine cmd/api/main.go
./engine > /dev/null 2>&1 &
SERVER_PID=$!

# Ensure the server is killed if the script is aborted early (Ctrl+C)
trap "kill -9 $SERVER_PID 2>/dev/null; lsof -ti:8080 | xargs kill -9 2>/dev/null || true" EXIT INT TERM

# Wait a moment for the server to bind to the port
sleep 2

# 5. Execute Concurrent Load Test (60 Seconds)
echo "🔥 Initiating 60-second concurrent load test..."
echo "📈 Target: 30,000+ Requests Per Second (15k Buys + 15k Sells)"

hey -m POST -D dummy-data/buy.json -T "application/json" -c 50 -q 300 -z 60s http://localhost:8080/api/v1/orders > /dev/null &
PID_BUY=$!

hey -m POST -D dummy-data/sell.json -T "application/json" -c 50 -q 300 -z 60s http://localhost:8080/api/v1/orders > /dev/null &
PID_SELL=$!

# Wait for both load tests to finish
wait $PID_BUY
wait $PID_SELL

echo "✅ Load test complete!"

# 6. Fetch and display metrics nicely
echo "=========================================================="
echo "📊 Final Performance Metrics"
echo "=========================================================="

curl -s http://localhost:8080/metrics | python3 -m json.tool

echo "=========================================================="
echo "🛑 Shutting down the Matching Engine..."

# Explicit teardown: Kill the specific background PID
kill -9 $SERVER_PID 2>/dev/null || true

# kill the process listening on port 8080
lsof -ti:8080 | xargs kill -9 2>/dev/null || true

echo "✅ Server stopped. Port 8080 is clear."
echo "🎉 Benchmark finished successfully."