package engine

import (
	"sync"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/metrics"
)

// Engine routes incoming requests to the correct symbol's processing goroutine.
type Engine struct {
	// We only need a mutex to protect the map when adding a brand-new symbol.
	// Once the symbol exists, sending to its channel is lock-free.
	mu             sync.RWMutex
	routers        map[string]chan interface{}
	GlobalOrderMap sync.Map // Maps orderID (string) to symbol (string)
	Metrics        *metrics.GlobalMetrics
}

func NewEngine() *Engine {
	return &Engine{
		routers: make(map[string]chan interface{}),
	}
}

// getOrSpawnRouter returns the channel for a symbol, creating it if it doesn't exist.
func (e *Engine) getOrSpawnRouter(symbol string) chan interface{} {
	// Fast path: Read lock to check if it exists
	e.mu.RLock()
	ch, exists := e.routers[symbol]
	e.mu.RUnlock()

	if exists {
		return ch
	}

	// Slow path: Write lock to create it
	e.mu.Lock()
	defer e.mu.Unlock()

	// Double-check in case another thread created it while we waited for the lock
	if ch, exists = e.routers[symbol]; exists {
		return ch
	}

	// Buffer of 10,000 to absorb massive traffic spikes without blocking the API layer
	ch = make(chan interface{}, 10000)
	e.routers[symbol] = ch

	// Spawn the dedicated worker goroutine for this symbol
	ob := NewOrderBook(symbol)
	go processSymbol(ob, ch, e.Metrics)

	return ch
}

// SubmitOrder routes an order to the correct symbol's worker.
func (e *Engine) SubmitOrder(req OrderRequest) {
	ch := e.getOrSpawnRouter(req.Order.Symbol)
	ch <- req
}

// CancelOrder routes a cancel request to the correct symbol's worker.
func (e *Engine) CancelOrder(symbol string, req CancelRequest) {
	ch := e.getOrSpawnRouter(symbol)
	ch <- req
}

// GetOrderBook routes a snapshot request to the correct symbol's worker.
func (e *Engine) GetOrderBook(symbol string, req SnapshotRequest) {
	ch := e.getOrSpawnRouter(symbol)
	ch <- req
}
