package model

import "github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/types"

type PriceLevel struct {
	Price types.Price

	Orders []*Order
}