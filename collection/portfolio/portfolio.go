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
		cash:   0,
	}
	return &port
}
func (port *Portfolio) Insert(h *instrument.Holding, t instrument.Tick) error {
	return port.active.InsertUpdate(h, t)
}

func (port *Portfolio) Delete(key string) error {
	return port.active.Delete(key)
}

func (port *Portfolio) Update(t instrument.Tick) error {
	return port.active.Update(t)
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
