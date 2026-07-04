package model

import (
	"time"

	"github.com/google/uuid"

	"container/list"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/types"
)

type Order struct {
	ID string `json:"id"`

	Symbol string `json:"symbol"`

	Side types.OrderSide `json:"side"`

	Type types.OrderType `json:"type"`

	Price types.Price `json:"price"`

	Quantity int64 `json:"quantity"`

	FilledQuantity int64 `json:"filled_quantity"`

	Status types.OrderStatus `json:"status"`

	Timestamp int64 `json:"timestamp"`

	QueueElement *list.Element `json:"-"`
}

func NewOrder(
	symbol string,
	side types.OrderSide,
	orderType types.OrderType,
	price types.Price,
	quantity int64,
) *Order {

	return &Order{
		ID:             uuid.NewString(),
		Symbol:         symbol,
		Side:           side,
		Type:           orderType,
		Price:          price,
		Quantity:       quantity,
		FilledQuantity: 0,
		Status:         types.StatusAccepted,
		Timestamp:      time.Now().UnixMilli(),
	}
}

func (o *Order) RemainingQuantity() int64 {
	return o.Quantity - o.FilledQuantity
}
