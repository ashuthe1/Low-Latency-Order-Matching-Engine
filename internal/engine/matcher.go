package engine

import (
	"errors"
	"time"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/models"
	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/google/uuid"
)

var ErrInsufficientLiquidity = errors.New("insufficient liquidity")

// ProcessOrder is the main entry point for an incoming order.
// It returns a slice of generated Trades and an error if the order is rejected.
func (ob *OrderBook) ProcessOrder(order *models.Order) ([]models.Trade, error) {
	if order.Type == models.Market {
		if !ob.hasSufficientLiquidity(order.Side, order.Quantity) {
			order.Status = models.Rejected
			return nil, ErrInsufficientLiquidity
		}
	}

	var trades []models.Trade

	// Determine which side of the book to match against
	var oppositeBook *redblacktree.Tree
	if order.Side == models.Buy {
		oppositeBook = ob.Asks
	} else {
		oppositeBook = ob.Bids
	}

	// Match while the order has remaining quantity and the opposite book is not empty
	for order.FilledQty < order.Quantity && oppositeBook.Size() > 0 {
		var bestPriceNode *redblacktree.Node
		if order.Side == models.Buy {
			bestPriceNode = oppositeBook.Left() // Lowest Ask
		} else {
			bestPriceNode = oppositeBook.Left() // Highest Bid (Tree is descending)
		}

		bestPriceLevel := bestPriceNode.Value.(*PriceLevel)
		bestPrice := bestPriceLevel.Price

		// Check if prices cross for Limit Orders
		if order.Type == models.Limit {
			if order.Side == models.Buy && order.Price < bestPrice {
				break // Buy price is lower than the best ask, no match
			}
			if order.Side == models.Sell && order.Price > bestPrice {
				break // Sell price is higher than the best bid, no match
			}
		}

		// Execute against the price level
		levelTrades := ob.matchAtPriceLevel(order, bestPriceLevel)
		trades = append(trades, levelTrades...)

		// If the price level is depleted, remove it from the tree
		if bestPriceLevel.Volume == 0 {
			oppositeBook.Remove(bestPrice)
		}
	}

	// Update order status based on fill amount
	if order.FilledQty == order.Quantity {
		order.Status = models.Filled
	} else if order.FilledQty > 0 {
		order.Status = models.PartialFill
	} else {
		order.Status = models.Accepted
	}

	// If it's a Limit order and not fully filled, it rests in the book
	if order.Type == models.Limit && order.FilledQty < order.Quantity {
		ob.AddOrder(order)
	}

	return trades, nil
}

// matchAtPriceLevel iterates through the FIFO queue of a price level and generates trades.
func (ob *OrderBook) matchAtPriceLevel(incoming *models.Order, level *PriceLevel) []models.Trade {
	var trades []models.Trade

	// Iterate through resting orders at this price
	for i := 0; i < len(level.Orders); i++ {
		restingOrder := level.Orders[i]

		// Skip lazy-deleted (cancelled) orders
		if restingOrder.Status == models.Cancelled {
			continue
		}

		incomingRemaining := incoming.Quantity - incoming.FilledQty
		restingRemaining := restingOrder.Quantity - restingOrder.FilledQty

		// Determine trade quantity (minimum of the two remainings)
		tradeQty := incomingRemaining
		if restingRemaining < tradeQty {
			tradeQty = restingRemaining
		}

		// Generate trade (Price is ALWAYS the resting order's price)
		trade := models.Trade{
			TradeID:   uuid.New().String(), // Ensure you run: go get github.com/google/uuid
			OrderID:   incoming.ID,         // API will map this trade to the aggressor
			Price:     level.Price,
			Quantity:  tradeQty,
			Timestamp: time.Now().UnixMilli(),
		}
		trades = append(trades, trade)

		// Update Quantities
		incoming.FilledQty += tradeQty
		restingOrder.FilledQty += tradeQty
		level.ReduceVolume(tradeQty)

		// Update resting order status
		if restingOrder.FilledQty == restingOrder.Quantity {
			restingOrder.Status = models.Filled
			// Lazy cleanup: we don't slice the array here to avoid O(N) shifts.
			// We just let the level's total Volume hit 0, which removes the whole node later.
		}

		// If incoming order is completely filled, we can stop matching
		if incoming.FilledQty == incoming.Quantity {
			break
		}
	}

	// Cleanup depleted orders from the front of the queue to free memory
	// O(K) where K is the number of filled orders at the front.
	head := 0
	for head < len(level.Orders) && level.Orders[head].Status == models.Filled {
		head++
	}
	if head > 0 {
		// Shift slice pointer forward to garbage collect filled orders
		level.Orders = level.Orders[head:]
	}

	return trades
}

// hasSufficientLiquidity validates if a market order can be fully filled.
func (ob *OrderBook) hasSufficientLiquidity(side models.Side, requiredQty int64) bool {
	var tree *redblacktree.Tree
	if side == models.Buy {
		tree = ob.Asks
	} else {
		tree = ob.Bids
	}

	var availableQty int64 = 0
	iterator := tree.Iterator()
	
	// Iterate through the tree until we have enough volume
	for iterator.Next() {
		level := iterator.Value().(*PriceLevel)
		availableQty += level.Volume
		if availableQty >= requiredQty {
			return true
		}
	}
	return false
}