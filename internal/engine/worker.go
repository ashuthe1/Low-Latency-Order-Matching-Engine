package engine

import (
	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/metrics"
)

// processSymbol is the isolated event loop for a single OrderBook.
// It reads from the channel and executes operations sequentially.
func processSymbol(ob *OrderBook, ch <-chan interface{}, m *metrics.GlobalMetrics) {
	for msg := range ch {
		switch req := msg.(type) {

		case OrderRequest:
			trades, err := ob.ProcessOrder(req.Order)

			if len(trades) > 0 && m != nil {
				m.TradesExecuted.Add(uint64(len(trades)))
				m.OrdersMatched.Add(1)
			}

			req.ResponseChan <- OrderResponse{
				Trades: trades,
				Error:  err,
			}

		case CancelRequest:
			err := ob.CancelOrder(req.OrderID)
			req.ResponseChan <- err

		case SnapshotRequest:
			// We must generate the snapshot inside this thread to avoid race conditions
			// when reading the Red-Black Trees.
			req.ResponseChan <- generateSnapshot(ob, req.Depth)
		}
	}
}

// generateSnapshot safely extracts the top N levels of the book for the API.
func generateSnapshot(ob *OrderBook, depth int) OrderBookSnapshot {
	snap := OrderBookSnapshot{
		Bids: make([]PriceLevelSnapshot, 0, depth),
		Asks: make([]PriceLevelSnapshot, 0, depth),
	}

	// Grab top Bids
	bidIter := ob.Bids.Iterator()
	count := 0
	for bidIter.Next() && count < depth {
		level := bidIter.Value().(*PriceLevel)
		if level.Volume > 0 { // Ignore lazy-deleted empty levels
			snap.Bids = append(snap.Bids, PriceLevelSnapshot{Price: level.Price, Volume: level.Volume})
			count++
		}
	}

	// Grab top Asks
	askIter := ob.Asks.Iterator()
	count = 0
	for askIter.Next() && count < depth {
		level := askIter.Value().(*PriceLevel)
		if level.Volume > 0 {
			snap.Asks = append(snap.Asks, PriceLevelSnapshot{Price: level.Price, Volume: level.Volume})
			count++
		}
	}

	return snap
}
