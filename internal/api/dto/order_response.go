package dto

import (
	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/model"
	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/types"
)

type SubmitOrderResponse struct {
	OrderID string `json:"order_id"`

	Status types.OrderStatus `json:"status"`

	FilledQuantity int64 `json:"filled_quantity,omitempty"`

	RemainingQuantity int64 `json:"remaining_quantity,omitempty"`

	Trades []model.Trade `json:"trades,omitempty"`

	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}