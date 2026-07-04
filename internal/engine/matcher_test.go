package engine

import (
	"testing"
	"time"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/models"
)

// Helper to quickly generate orders for tests
func newTestOrder(id, side, orderType string, price, qty int64) *models.Order {
	return &models.Order{
		ID:        id,
		Symbol:    "AAPL",
		Side:      models.Side(side),
		Type:      models.OrderType(orderType),
		Price:     price,
		Quantity:  qty,
		Timestamp: time.Now().UnixMilli(),
	}
}

func TestExample1_SimpleFullMatch(t *testing.T) {
	ob := NewOrderBook("AAPL")

	// Initial State
	sellOrder := newTestOrder("order-001", "SELL", "LIMIT", 15050, 1000)
	buyOrder := newTestOrder("order-002", "BUY", "LIMIT", 15045, 500)

	ob.AddOrder(sellOrder)
	ob.AddOrder(buyOrder)

	// New Order Arrives
	newBuy := newTestOrder("order-NEW", "BUY", "LIMIT", 15050, 500)
	trades, err := ob.ProcessOrder(newBuy)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(trades))
	}

	if trades[0].Price != 15050 || trades[0].Quantity != 500 {
		t.Errorf("Trade executed at wrong price/qty: %+v", trades[0])
	}

	if newBuy.Status != models.Filled {
		t.Errorf("New buy order should be FILLED, got %s", newBuy.Status)
	}
}

func TestExample2_MultiplePriceLevels(t *testing.T) {
	ob := NewOrderBook("AAPL")

	ob.AddOrder(newTestOrder("order-003", "SELL", "LIMIT", 15050, 300))
	ob.AddOrder(newTestOrder("order-004", "SELL", "LIMIT", 15052, 400))
	ob.AddOrder(newTestOrder("order-005", "SELL", "LIMIT", 15055, 600))

	newBuy := newTestOrder("order-NEW", "BUY", "LIMIT", 15053, 800)
	trades, _ := ob.ProcessOrder(newBuy)

	if len(trades) != 2 {
		t.Fatalf("Expected 2 trades, got %d", len(trades))
	}

	if trades[0].Quantity != 300 || trades[0].Price != 15050 {
		t.Errorf("First trade incorrect: %+v", trades[0])
	}
	if trades[1].Quantity != 400 || trades[1].Price != 15052 {
		t.Errorf("Second trade incorrect: %+v", trades[1])
	}

	if newBuy.Status != models.PartialFill {
		t.Errorf("Order should be PARTIAL_FILL, got %s", newBuy.Status)
	}
	if newBuy.FilledQty != 700 {
		t.Errorf("Expected 700 filled, got %d", newBuy.FilledQty)
	}
}

func TestExample3_TimePriorityFIFO(t *testing.T) {
	ob := NewOrderBook("AAPL")

	// Add three identical orders, order matters
	o1 := newTestOrder("order-007", "SELL", "LIMIT", 15050, 200)
	o2 := newTestOrder("order-008", "SELL", "LIMIT", 15050, 300)
	o3 := newTestOrder("order-009", "SELL", "LIMIT", 15050, 400)

	ob.AddOrder(o1)
	ob.AddOrder(o2)
	ob.AddOrder(o3)

	newBuy := newTestOrder("order-NEW", "BUY", "LIMIT", 15050, 500)
	trades, _ := ob.ProcessOrder(newBuy)

	if len(trades) != 2 {
		t.Fatalf("Expected 2 trades, got %d", len(trades))
	}

	if o1.Status != models.Filled {
		t.Errorf("First order should be completely filled")
	}
	if o2.Status != models.Filled {
		t.Errorf("Second order should be completely filled")
	}
	if o3.FilledQty != 0 {
		t.Errorf("Third order should be untouched")
	}
}

func TestExample4_MarketOrder(t *testing.T) {
	ob := NewOrderBook("AAPL")

	ob.AddOrder(newTestOrder("order-010", "SELL", "LIMIT", 15050, 200))
	ob.AddOrder(newTestOrder("order-011", "SELL", "LIMIT", 15052, 300))
	ob.AddOrder(newTestOrder("order-012", "SELL", "LIMIT", 15055, 400))

	marketBuy := newTestOrder("order-NEW", "BUY", "MARKET", 0, 600)
	trades, _ := ob.ProcessOrder(marketBuy)

	if len(trades) != 3 {
		t.Fatalf("Expected 3 trades, got %d", len(trades))
	}
	if marketBuy.Status != models.Filled {
		t.Errorf("Market order should be filled")
	}
}

func TestExample5_InsufficientLiquidity(t *testing.T) {
	ob := NewOrderBook("AAPL")

	ob.AddOrder(newTestOrder("order-013", "SELL", "LIMIT", 15050, 100))

	marketBuy := newTestOrder("order-NEW", "BUY", "MARKET", 0, 500)
	trades, err := ob.ProcessOrder(marketBuy)

	if err != ErrInsufficientLiquidity {
		t.Errorf("Expected Insufficient Liquidity error, got %v", err)
	}
	if len(trades) > 0 {
		t.Errorf("No trades should have executed")
	}
	if marketBuy.Status != models.Rejected {
		t.Errorf("Market order should be REJECTED")
	}
}
