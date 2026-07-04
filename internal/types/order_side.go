package types

type OrderSide string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

func (s OrderSide) Valid() bool {
	return s == Buy || s == Sell
}
