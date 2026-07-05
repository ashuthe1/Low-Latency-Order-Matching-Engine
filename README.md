# High-Performance Order Matching Engine

A low-latency, highly concurrent order matching engine written in Go. This system implements strict Price-Time (FIFO) priority matching for limit and market orders, capable of sustaining massive throughput with sub-millisecond tail latencies.

## 🚀 Quick Start

### Prerequisites
* Go 1.22 or higher
* `make` (for build orchestration)
* `hey` (for load testing). If you don't have it installed, run:
  ```bash
  go install [github.com/rakyll/hey@latest](https://github.com/rakyll/hey@latest)

```

### Build and Run

Start the API server on port `8080` using `make`:

```bash
make run

```

*(Alternatively: `go run cmd/api/main.go`)*

### Testing

Run the core matching logic unit and validation tests:

```bash
make test

```

To run tests with Go's race detector enabled to verify complete thread safety:

```bash
make race

```

---

## 🏎️ Benchmarking & Load Testing

The system is architected to survive brutal execution spikes. You can evaluate engine thresholds via the automated suite or manual connection piping.

### Automated Benchmark Suite (Recommended)

The throughput computation exposed by the `/metrics` endpoint dynamically divides total processed orders by the **total server uptime**. Leaving the server running idle before or after a load sequence dilutes the reported throughput average.

To isolate exact peak capabilities, use the automated suite. It builds the target binary, initializes isolation parameters, maps data buffers, fires a simultaneous 100,000 TPS workload (50k Buys + 50k Sells) for 60 seconds using `hey`, yields structural JSON tracking metrics, and tears down the infrastructure cleanly.

```bash
make benchmark
# Or execute manually: ./benchmark.sh

```

### Manual Load Testing

If running scenarios step-by-step, ensure the server is active via `make run`.

1. **Obtain Test Payloads:** Running the `./benchmark.sh` pipeline once auto-generates a `dummy-data/` hierarchy containing pre-formatted `buy.json` and `sell.json` payloads.
2. **Execute Cross-Load Streams:** Pipe overlapping BUY and SELL sequences in parallel to simulate real market depth conditions. This command establishes 100 concurrent pipes scaling up to 100,000 requests per second:

```bash
# Launch BUY allocations to background workers
hey -m POST -D dummy-data/buy.json -T "application/json" -c 100 -q 500 -z 60s http://localhost:8080/api/v1/orders &

# Launch matching SELL allocations synchronously
hey -m POST -D dummy-data/sell.json -T "application/json" -c 100 -q 500 -z 60s http://localhost:8080/api/v1/orders

```

3. **Inspect Application State:** Query the live tracking counters immediately following stream convergence:

```bash
curl http://localhost:8080/metrics

```

---

## 🏗️ Architecture & Approach

To satisfy the demanding >30,000 TPS limits without hitting synchronization bottlenecks, this engine rejects traditional paradigms that wrap core structures in global `sync.RWMutex` locks, which inevitably degrade into severe critical-section contention.

Instead, this system leans natively on the **Actor Model** utilizing Go primitives:

1. **Per-Symbol Worker Sharding:** Isolated symbols (e.g., AAPL, TSLA) are bound exclusively to an unshared worker goroutine communicating via isolated 10,000-slot buffered channels.
2. **Lock-Free Execution Planes:** The API routing layer ingests JSON commands, validates limits, and forwards work directly to the specific symbol thread pool. Because exactly one thread reads and mutates an asset's book state, **internal order book matching completely drops mutex protections**, eliminating memory contention vectors and deadlock possibilities.
3. **Strict Integer Mathematics:** Every monetary value, execution tier, and structural capacity metric uses `int64` fixed-point math exclusively. Floating-point primitives are banned, ensuring 100% calculation precision and optimal CPU register arithmetic efficiency.

### Data Structures

* **Price Level Queues:** Array-backed contiguous memory slices facilitate O(1) FIFO scheduling loops. These blocks leverage hardware cache localization properties perfectly for O(1) appends and O(1) front-pops.
* **Order Books:** Bid and Ask ranges are organized across balanced Red-Black Trees. This guarantees strict $O(\log N)$ overhead for deep tracking injections and immediate $O(1)$ visibility over the current market spreads.
* **Global Registry Mapping:** A concurrently safe global `sync.Map` bridges incoming `OrderID` identifiers directly back to parent symbol channels to maintain $O(1)$ time complexity bounds during immediate client-side cancellations.

---

## 📊 Performance Results

*Validations derived on a 4-core, 16GB development machine using `benchmark.sh`.*

| Metric | Target Requirement | Actual Engine Result |
| --- | --- | --- |
| **Throughput (TPS)** | ≥ 30,000 orders/sec | **~84,720 orders/second** |
| **Latency (p50)** | ≤ 10 ms | **0.009 ms** |
| **Latency (p99)** | ≤ 50 ms | **2.373 ms** |
| **Latency (p999)** | ≤ 100 ms | **5.255 ms** |
| **Correctness Profile** | 100% Valid | **100% Verified** (0 race flags) |

*Telemetry tracked using high-precision atomic state arrays and structural clock delta records to drop parsing noise during profiling.*

---

### 🧰 Postman Collection
For easy manual testing and API exploration, a complete Postman collection is included with this repository. Simply navigate to the `postman/` directory and import the exported JSON file into your Postman workspace to instantly interact with all available endpoints.

---

## ⚠️ Assumptions & Limitations

* **Self-Match Permissibility:** In absolute alignment with evaluation bounds, cross-client account prevention checks are bypassed; orders matching matching accounts execute fields identically.
* **Volatile Persistence Boundary:** Storage mappings exist entirely in-memory to prioritize maximal throughput performance. Unsaved order context clears on engine restart.
* **Optimized Cancel Flags:** Dropping orders skips expensive $O(N)$ memory shifting sweeps. Elements flag as cancelled in $O(1)$ time bounds and clear dynamically when hit by matching waves.

## 🚀 Future Improvements

1. WAL: Implement Write-Ahead-Log flow, so that if system crashes the orderbook or any data related to them don't get lost.
2. **Allocation pooling via `sync.Pool`:** Reusing volatile transient components like `Order` and `Trade` structs reduces Go Garbage Collector memory scans, flatlining the microsecond p999 latency tail during continuous workload spikes.
3. **Server Sent Event(SSE) Snapshot Streaming:** Exposing subscription structures via real-time SSE listeners can deliver instant match ticks and depth adjustments down to high-frequency consumers.