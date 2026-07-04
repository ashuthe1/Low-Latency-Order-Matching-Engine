package engine

import (
	"errors"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/models"
	"github.com/emirpasic/gods/trees/redblacktree"
)

// Custom Comparators for the Red-Black Trees
// We must use int64 assertions since prices are stored as int64.

// ascComparator sorts Asks lowest-to-highest
func ascComparator(a, b interface{}) int {
	aPrice, bPrice := a.(int64), b.(int64)
	switch {
	case aPrice > bPrice:
		return 1
	case aPrice < bPrice:
		return -1
	default:
		return 0
	}
}

// descComparator sorts Bids highest-to-lowest
func descComparator(a, b interface{}) int {
	aPrice, bPrice := a.(int64), b.(int64)
	switch {
	case aPrice < bPrice:
		return 1
	case aPrice > bPrice:
		return -1
	default:
		return 0
	}
}

// OrderBook manages all active limit orders for a single symbol.
type OrderBook struct {
	Symbol       string
	Bids         *redblacktree.Tree // Tree of *PriceLevel (Descending)
	Asks         *redblacktree.Tree // Tree of *PriceLevel (Ascending)
	ActiveOrders map[string]*models.Order
}

// NewOrderBook initializes the OrderBook for a specific symbol.
func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol:       symbol,
		Bids:         redblacktree.NewWith(descComparator),
		Asks:         redblacktree.NewWith(ascComparator),
		ActiveOrders: make(map[string]*models.Order),
	}
}

// AddOrder places a LIMIT order into the book.
// Note: MARKET orders do not rest in the book, so they bypass this.
func (ob *OrderBook) AddOrder(order *models.Order) {
	ob.ActiveOrders[order.ID] = order

	var tree *redblacktree.Tree
	if order.Side == models.Buy {
		tree = ob.Bids
	} else {
		tree = ob.Asks
	}

	// O(log N) lookup/insert for the price level
	node, found := tree.Get(order.Price)
	var level *PriceLevel

	if found {
		level = node.(*PriceLevel)
	} else {
		level = NewPriceLevel(order.Price)
		tree.Put(order.Price, level)
	}

	// O(1) append to the price level queue
	level.Append(order)
}

// CancelOrder handles O(1) cancellation via lazy deletion.
func (ob *OrderBook) CancelOrder(orderID string) error {
	order, exists := ob.ActiveOrders[orderID]
	if !exists {
		return errors.New("order not found")
	}
	if order.Status == models.Filled || order.Status == models.Cancelled {
		return errors.New("cannot cancel order: already filled or cancelled")
	}

	// Mark as cancelled. The matching engine will skip it when evaluating queues.
	order.Status = models.Cancelled
	// delete(ob.ActiveOrders, orderID)

	// We still need to decrement the volume from the price level (O(log N) operation)
	var tree *redblacktree.Tree
	if order.Side == models.Buy {
		tree = ob.Bids
	} else {
		tree = ob.Asks
	}

	node, found := tree.Get(order.Price)
	if found {
		level := node.(*PriceLevel)
		remainingQty := order.Quantity - order.FilledQty
		level.ReduceVolume(remainingQty)

		// If volume reaches 0, we can safely prune the node from the tree
		if level.Volume <= 0 {
			tree.Remove(order.Price)
		}
	}

	return nil
}
