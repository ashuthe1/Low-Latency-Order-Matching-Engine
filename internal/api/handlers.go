package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/engine"
	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/models"
	"github.com/google/uuid"
)

type API struct {
	Engine        *engine.Engine
	OrderToSymbol *sync.Map // Maps OrderID -> Symbol for fast routing
}

func NewAPI(eng *engine.Engine) *API {
	return &API{
		Engine:        eng,
		OrderToSymbol: &sync.Map{},
	}
}

// respondJSON is a high-performance helper to write JSON responses
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// respondError standardizes error responses
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// POST /api/v1/orders
func (a *API) HandleSubmitOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	// Basic Validation
	if order.Quantity <= 0 {
		respondError(w, http.StatusBadRequest, "Invalid order: quantity must be positive")
		return
	}
	if order.Type == models.Limit && order.Price <= 0 {
		respondError(w, http.StatusBadRequest, "Invalid order: limit orders require a positive price")
		return
	}

	// Enrich Order
	order.ID = uuid.New().String()
	order.Timestamp = time.Now().UnixMilli()
	order.Status = models.Accepted

	// Save routing lookup
	a.OrderToSymbol.Store(order.ID, order.Symbol)

	// Send to Engine and wait for response
	respChan := make(chan engine.OrderResponse, 1)
	a.Engine.SubmitOrder(engine.OrderRequest{
		Order:        &order,
		ResponseChan: respChan,
	})

	result := <-respChan

	// Handle Insufficient Liquidity for Market Orders
	if result.Error == engine.ErrInsufficientLiquidity {
		a.OrderToSymbol.Delete(order.ID) // cleanup
		respondError(w, http.StatusBadRequest, result.Error.Error())
		return
	}

	// Format Response based on fill status
	response := map[string]interface{}{
		"order_id": order.ID,
		"status":   order.Status,
	}

	if order.Status == models.Filled {
		response["filled_quantity"] = order.FilledQty
		response["trades"] = result.Trades
		respondJSON(w, http.StatusOK, response)
	} else if order.Status == models.PartialFill {
		response["filled_quantity"] = order.FilledQty
		response["remaining_quantity"] = order.Quantity - order.FilledQty
		response["trades"] = result.Trades
		respondJSON(w, http.StatusAccepted, response)
	} else {
		response["message"] = "Order added to book"
		respondJSON(w, http.StatusCreated, response)
	}
}

// DELETE /api/v1/orders/{order_id}
func (a *API) HandleCancelOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("order_id") // Go 1.22 routing feature

	// O(1) lookup to find which symbol thread to route to
	symbolInter, exists := a.OrderToSymbol.Load(orderID)
	if !exists {
		respondError(w, http.StatusNotFound, "Order not found")
		return
	}
	symbol := symbolInter.(string)

	respChan := make(chan error, 1)
	a.Engine.CancelOrder(symbol, engine.CancelRequest{
		OrderID:      orderID,
		ResponseChan: respChan,
	})

	err := <-respChan
	if err != nil {
		if err.Error() == "order not found" {
			respondError(w, http.StatusNotFound, err.Error())
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"order_id": orderID,
		"status":   "CANCELLED",
	})
}

// GET /api/v1/orderbook/{symbol}
func (a *API) HandleGetOrderBook(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")

	// Parse depth query param (default to 10)
	depth := 10
	if depthStr := r.URL.Query().Get("depth"); depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil && d > 0 {
			depth = d
		}
	}

	respChan := make(chan engine.OrderBookSnapshot, 1)
	a.Engine.GetOrderBook(symbol, engine.SnapshotRequest{
		Depth:        depth,
		ResponseChan: respChan,
	})

	snapshot := <-respChan

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"symbol":    symbol,
		"timestamp": time.Now().UnixMilli(),
		"bids":      snapshot.Bids,
		"asks":      snapshot.Asks,
	})
}
