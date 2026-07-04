package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/types"
)

type Trade struct {
	ID string `json:"trade_id"`

	BuyOrderID string `json:"buy_order_id"`

	SellOrderID string `json:"sell_order_id"`

	Symbol string `json:"symbol"`

	Price types.Price `json:"price"`

	Quantity int64 `json:"quantity"`

	Timestamp int64 `json:"timestamp"`
}

func NewTrade(
	buyID,
	sellID,
	symbol string,
	price types.Price,
	qty int64,
) Trade {

	return Trade{
		ID:          uuid.NewString(),
		BuyOrderID:  buyID,
		SellOrderID: sellID,
		Symbol:      symbol,
		Price:       price,
		Quantity:    qty,
		Timestamp:   time.Now().UnixMilli(),
	}
}