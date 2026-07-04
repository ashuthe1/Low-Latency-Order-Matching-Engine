package engine

import "github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/models"

// OrderResponse is sent back to the API layer after matching.
type OrderResponse struct {
	Trades []models.Trade
	Error  error
}

// OrderRequest carries a new order and a channel to receive the result.
type OrderRequest struct {
	Order        *models.Order
	ResponseChan chan<- OrderResponse
}

// CancelRequest carries an order ID to cancel and a channel for the result.
type CancelRequest struct {
	OrderID      string
	ResponseChan chan<- error
}

// OrderBookSnapshot represents a point-in-time view of the book for the API.
type OrderBookSnapshot struct {
	Bids []PriceLevelSnapshot `json:"bids"`
	Asks []PriceLevelSnapshot `json:"asks"`
}

type PriceLevelSnapshot struct {
	Price  int64 `json:"price"`
	Volume int64 `json:"volume"`
}

// SnapshotRequest asks the engine for a copy of the current order book state.
type SnapshotRequest struct {
	Depth        int
	ResponseChan chan<- OrderBookSnapshot
}

type StatusResponse struct {
	Order *models.Order
	Error error
}

type StatusRequest struct {
	OrderID      string
	ResponseChan chan<- StatusResponse
}
