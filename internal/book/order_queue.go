package book

import (
	"container/list"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/model"
)

type OrderQueue struct {
	list *list.List
}

func NewOrderQueue() *OrderQueue {
	return &OrderQueue{
		list: list.New(),
	}
}

func (q *OrderQueue) Push(order *model.Order) *list.Element {
	return q.list.PushBack(order)
}

func (q *OrderQueue) Front() *model.Order {
	if q.list.Len() == 0 {
		return nil
	}

	return q.list.Front().Value.(*model.Order)
}

func (q *OrderQueue) Pop() *model.Order {

	if q.list.Len() == 0 {
		return nil
	}

	element := q.list.Front()

	q.list.Remove(element)

	return element.Value.(*model.Order)
}

func (q *OrderQueue) Remove(element *list.Element) {

	if element == nil {
		return
	}

	q.list.Remove(element)
}

func (q *OrderQueue) Len() int {
	return q.list.Len()
}

func (q *OrderQueue) Empty() bool {
	return q.list.Len() == 0
}
