package models

// Side represents the side of the order (BUY or SELL)
type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

// OrderType represents whether the order is LIMIT or MARKET
type OrderType string

const (
	Limit  OrderType = "LIMIT"
	Market OrderType = "MARKET"
)

// OrderStatus tracks the lifecycle of an order in the system
type OrderStatus string

const (
	Accepted    OrderStatus = "ACCEPTED"
	PartialFill OrderStatus = "PARTIAL_FILL"
	Filled      OrderStatus = "FILLED"
	Cancelled   OrderStatus = "CANCELLED"
	Rejected    OrderStatus = "REJECTED"
)