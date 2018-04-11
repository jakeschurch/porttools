package portfolio

import (
	"errors"
	"sync"

	"github.com/jakeschurch/porttools/collection"
	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/utils"
)

var (
	// ErrNegativeCash indicates negative cash balance
	ErrNegativeCash = errors.New("Insufficient Funds")
)

// Portfolio struct refer to the aggregation of positions traded by a broker.
type Portfolio struct {
	active *collection.HoldingList
	mu     sync.RWMutex
	cash   utils.Amount
}

// New creates a new instance of a Portfolio struct.
func New() *Portfolio {
	port := Portfolio{
		active: collection.NewHoldingList(),
	}
	return &port
}

func (port *Portfolio) Insert(h *instrument.Holding, q instrument.Quote) error {
	return port.active.InsertUpdate(h, q)
}

func (port *Portfolio) Delete(key string) error {
	return port.active.Delete(key)
}

func (port *Portfolio) Update(q instrument.Quote) error {
	return port.active.Update(q)
}

// Pop returns a LinkedNode struct, will return element at head.next or tail
// position depending on the CostMethod specified.
func (port *Portfolio) Pop(key string, c utils.CostMethod) (*collection.LinkedNode, error) {
	var linkedList *collection.LinkedList
	var err error

	if linkedList, err = port.active.Get(key); err != nil {
		return nil, err
	}
	return linkedList.Pop(c), nil
}

// Peek will return an element from a Linked List, depending on the key given
// as well as the cost method. Will return nil if nothing found.
func (port *Portfolio) Peek(key string, c utils.CostMethod) *collection.LinkedNode {
	var err error
	var list *collection.LinkedList

	if list, err = port.active.Get(key); err != nil {
		return nil
	}
	return list.Peek(c)
}

func (port *Portfolio) UpdateCash(delta utils.Amount) {
	port.mu.Lock()
	port.cash += delta
	port.mu.Unlock()
}

func (port *Portfolio) GetList(key string) (*collection.LinkedList, error) {
	var list *collection.LinkedList
	var err error

	port.mu.RLock()
	list, err = port.active.Get(key)
	port.mu.RUnlock()

	if err != nil {
		return nil, err
	}
	return list, nil
}
