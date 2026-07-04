package engine

import "github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/models"

// PriceLevel represents a queue of orders at a specific price.
// It maintains time-priority (FIFO) via the slice ordering.
type PriceLevel struct {
	Price  int64
	Orders []*models.Order
	Volume int64 // Total available quantity at this price level
}

// NewPriceLevel initializes a price level with pre-allocated capacity
// to reduce garbage collection overhead during high load.
func NewPriceLevel(price int64) *PriceLevel {
	return &PriceLevel{
		Price:  price,
		Orders: make([]*models.Order, 0, 128),
		Volume: 0,
	}
}

// Append adds an order to the end of the queue O(1)
func (pl *PriceLevel) Append(order *models.Order) {
	pl.Orders = append(pl.Orders, order)
	pl.Volume += order.Quantity
}

// ReduceVolume is called when an order is partially filled or cancelled.
func (pl *PriceLevel) ReduceVolume(amount int64) {
	pl.Volume -= amount
}