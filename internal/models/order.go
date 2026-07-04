package models

type Order struct {
	ID        string      `json:"id"`
	Symbol    string      `json:"symbol"`
	Side      Side        `json:"side"`
	Type      OrderType   `json:"type"`
	Price     int64       `json:"price,omitempty"` // Omitempty for MARKET orders
	Quantity  int64       `json:"quantity"`
	FilledQty int64       `json:"filled_quantity"`
	Status    OrderStatus `json:"status"`
	Timestamp int64       `json:"timestamp"` // Unix milliseconds
}

type Trade struct {
	TradeID   string `json:"trade_id"`
	OrderID   string `json:"order_id"`
	Price     int64  `json:"price"`
	Quantity  int64  `json:"quantity"`
	Timestamp int64  `json:"timestamp"`
}
