package types

type OrderType string

const (
	LimitOrder  OrderType = "LIMIT"
	MarketOrder OrderType = "MARKET"
)

func (t OrderType) Valid() bool {
	return t == LimitOrder || t == MarketOrder
}