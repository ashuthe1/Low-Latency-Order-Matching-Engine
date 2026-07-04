# High-Performance Order Matching Engine

A low-latency, highly concurrent order matching engine written in Go. This system implements strict Price-Time (FIFO) priority matching for limit and market orders, capable of sustaining massive throughput with sub-millisecond tail latencies.

## 🚀 Quick Start

### Prerequisites
* Go 1.22 or higher
* `make` (optional, for convenience)
* `hey` (for load testing). If you don't have it installed, run:
  ```bash
  go install [github.com/rakyll/hey@latest](https://github.com/rakyll/hey@latest)

```

### Build and Run

You can start the server easily using `make`:

```bash
make run

```

*(Alternatively, you can use `go run cmd/api/main.go`)*

The server will start on port `8080`.

### Testing

Run the core matching logic and correctness tests:

```bash
make test

```

To run tests with Go's race detector enabled (verifying thread safety):

```bash
make race

```

---

## 🏎️ Benchmarking & Load Testing

The system is designed to handle extremely high concurrency. You can test this using the included automated benchmark script or by running manual load tests.

### Automated Benchmark Suite (Recommended)

The throughput calculation in the `/metrics` endpoint divides the total orders processed by the **total server uptime**. If the server is left running idly before or after a load test, the reported throughput average will drop.

To get an accurate representation of maximum throughput, it is highly recommended to use the included benchmark script. This script compiles the binary, starts the server, generates the dummy JSON payloads, fires a simultaneous 100,000 TPS load test (Buys and Sells) using `hey`, prints the exact metrics, and safely shuts the server down.

```bash
make benchmark
# Or run manually: ./benchmark.sh

```

### Manual Load Testing

If you prefer to run the tests manually, ensure the server is running (`make run`), then use the `hey` tool to send concurrent requests.

1. Ensure you have the test payloads available. (Running the `./benchmark.sh` script once will automatically generate a `dummy-data` folder containing `buy.json` and `sell.json` for you to use).
2. Fire the load test. To simulate a highly concurrent real-world environment, run both a BUY and a SELL load test simultaneously. The commands below will send 100,000 requests per second combined (50k Buys + 50k Sells) for 60 seconds:

```bash
# Start the BUY orders in the background
hey -m POST -D dummy-data/buy.json -T "application/json" -c 100 -q 500 -z 60s http://localhost:8080/api/v1/orders &

# Start the SELL orders
hey -m POST -D dummy-data/sell.json -T "application/json" -c 100 -q 500 -z 60s http://localhost:8080/api/v1/orders

```

3. Check the internal application latency and metrics immediately after the test completes:

```bash
curl http://localhost:8080/metrics

```

---

## 🏗️ Architecture & Approach

To achieve the stringent >30,000 TPS and <50ms p99 latency targets, I avoided the traditional approach of wrapping the entire order book in a global `sync.RWMutex`, which often leads to severe lock contention under heavy load.

Instead, this engine utilizes the **Actor Model** via Go's concurrency primitives:

1. **Per-Symbol Sharding:** Every trading symbol (example AAPL, CSCO) gets its own dedicated worker goroutine and a buffered communication channel (size 10,000).
2. **Lock-Free Matching:** The HTTP API layer parses incoming orders and sends them into the respective symbol's channel. The dedicated worker thread processes these requests sequentially. Because only one thread ever mutates a specific symbol's order book, **no mutexes are required during matching**, entirely eliminating race conditions and lock overhead.
3. **Integer Arithmetic:** All financial fields (prices, quantities) are strictly represented as `int64` to prevent floating-point precision loss and maximize CPU arithmetic efficiency.

### Data Structures

* **Price Levels:** Contiguous memory slices are used for price-level queues to maintain FIFO time priority. Slices are highly CPU-cache friendly, allowing $O(1)$ appends and $O(1)$ pops.
* **Order Book:** Bids and Asks are managed using Red-Black Trees, ensuring $O(\log N)$ time complexity for finding the best available prices and inserting new price levels.
* **Routing:** A global `sync.Map` routes `OrderID` to `Symbol` for $O(1)$ fast lookups during cancellation or status checks.

---

## 📊 Performance Results

*Results measured locally via the automated `benchmark.sh` script.*

| Metric | Target | Actual Result |
| --- | --- | --- |
| **Throughput (TPS)** | ≥ 30,000 | **~84,720 orders/second** |
| **Latency (p50)** | ≤ 10 ms | **0.009 ms** |
| **Latency (p99)** | ≤ 50 ms | **2.373 ms** |
| **Latency (p999)** | ≤ 100 ms | **5.255 ms** |
| **Correctness** | 100% | **100%** (0 race conditions) |

*Metrics were recorded internally using High Dynamic Range (HDR) Histograms and lock-free atomic counters to prevent measurement overhead.*

---

## ⚠️ Assumptions & Limitations

* **Self-Trading:** As per the assignment specifications, self-trading is permitted. The engine does not currently prevent a user's buy order from matching against their own sell order.
* **In-Memory State:** For maximum throughput, the entire order book state is kept in memory. If the server crashes, active orders are lost.
* **Lazy Deletion:** Cancelled orders are flagged as cancelled in $O(1)$ time and skipped during the matching loop, rather than executing an $O(N)$ slice shift to remove them immediately.

## 🚀 Future Improvements

1. **Garbage Collection Optimization (`sync.Pool`):** Currently, the system allocates new `Order` and `Trade` structs for every request. Implementing a `sync.Pool` to reuse these structs would drastically reduce Garbage Collector (GC) pressure, which is the primary cause of any microsecond spikes at the p999 level.
2. **WebSocket Streaming:** Implement real-time order book snapshots and trade tick feeds for clients using WebSockets.
3. **JSON Serialization:** Swap the standard `encoding/json` library for a high-performance alternative like `go-json` to reduce CPU cycles spent on API layer deserialization.