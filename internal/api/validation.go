package api

import (
	"errors"
	"strings"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/api/dto"
	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/types"
)

func ValidateOrderRequest(req dto.SubmitOrderRequest) error {

	req.Symbol = strings.TrimSpace(req.Symbol)

	if req.Symbol == "" {
		return errors.New("symbol is required")
	}

	if !req.Side.Valid() {
		return errors.New("invalid order side")
	}

	if !req.Type.Valid() {
		return errors.New("invalid order type")
	}

	if req.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	if req.Type == types.LimitOrder && !req.Price.Valid() {
		return errors.New("limit order requires positive price")
	}

	if req.Type == types.MarketOrder && req.Price != 0 {
		return errors.New("market order should not include price")
	}

	return nil
}
