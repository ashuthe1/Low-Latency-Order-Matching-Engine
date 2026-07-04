package dto

import "github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/types"

type SubmitOrderRequest struct {
	Symbol string `json:"symbol"`

	Side types.OrderSide `json:"side"`

	Type types.OrderType `json:"type"`

	Price types.Price `json:"price,omitempty"`

	Quantity int64 `json:"quantity"`
}