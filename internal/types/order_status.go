package types

type OrderStatus string

const (
	StatusAccepted    OrderStatus = "ACCEPTED"
	StatusPartialFill OrderStatus = "PARTIAL_FILL"
	StatusFilled      OrderStatus = "FILLED"
	StatusCancelled   OrderStatus = "CANCELLED"
)