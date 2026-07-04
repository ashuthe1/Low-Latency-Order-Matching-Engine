package book

import (
	"testing"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/model"
	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/types"
)

func createOrder(price int64) *model.Order {
	return model.NewOrder(
		"AAPL",
		types.Buy,
		types.LimitOrder,
		types.Price(price),
		100,
	)
}

func TestQueuePushPop(t *testing.T) {

	q := NewOrderQueue()

	o1 := createOrder(100)
	o2 := createOrder(101)
	o3 := createOrder(102)

	q.Push(o1)
	q.Push(o2)
	q.Push(o3)

	if q.Len() != 3 {
		t.Fatal("expected queue length 3")
	}

	if q.Pop() != o1 {
		t.Fatal("FIFO violated")
	}

	if q.Pop() != o2 {
		t.Fatal("FIFO violated")
	}

	if q.Pop() != o3 {
		t.Fatal("FIFO violated")
	}

	if !q.Empty() {
		t.Fatal("queue should be empty")
	}
}

func TestQueueFront(t *testing.T) {

	q := NewOrderQueue()

	o1 := createOrder(100)

	q.Push(o1)

	if q.Front() != o1 {
		t.Fatal("front mismatch")
	}
}

func TestQueueRemove(t *testing.T) {

	q := NewOrderQueue()

	o1 := createOrder(100)
	o2 := createOrder(101)
	o3 := createOrder(102)

	e1 := q.Push(o1)
	e2 := q.Push(o2)
	e3 := q.Push(o3)

	_ = e1
	_ = e3

	q.Remove(e2)

	if q.Len() != 2 {
		t.Fatal("remove failed")
	}

	if q.Pop() != o1 {
		t.Fatal("wrong order")
	}

	if q.Pop() != o3 {
		t.Fatal("wrong order")
	}
}

func TestEmptyQueue(t *testing.T) {

	q := NewOrderQueue()

	if q.Pop() != nil {
		t.Fatal("expected nil")
	}

	if q.Front() != nil {
		t.Fatal("expected nil")
	}
}
