package collection

import (
	"errors"
	"sync"

	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/instrument/holding"
	"github.com/jakeschurch/porttools/utils"
)

var (

	// ErrSliceExists indicates that a slice already exists
	ErrSliceExists = errors.New("Slice with ticker exists")

	// ErrNoSliceExists indicates that a slice already exists
	ErrNoSliceExists = errors.New("Slice with ticker does not exist")

	// ErrNegativeVolume indicates negative position volume balance
	ErrNegativeVolume = errors.New("Position Volume is less than 0")
)

// NewHoldingSlice returns a new holding slice.
func NewHoldingSlice() *HoldingSlice {
	holdingSlice := &HoldingSlice{
		len:         0,
		totalVolume: 0,
		holdings:    make([]*holding.Holding, 0),
	}
	return holdingSlice
}

// HoldingSlice is a slice that holds pointer values to holding.Holding type variables
type HoldingSlice struct {
	sync.RWMutex
	len         int
	holdings    []*holding.Holding
	totalVolume utils.Amount
}

// AddNew adds a new holding to the holdings slice.
func (slice *HoldingSlice) AddNew(newHolding *holding.Holding) error {
	// slice.Lock()
	slice.holdings = append(slice.holdings, newHolding)
	slice.len++
	// slice.Unlock()

	return slice.ApplyDelta(newHolding.Volume)
}

// UpdateMetrics update holdings data based on most recent tick data.
func (slice *HoldingSlice) UpdateMetrics(tick instrument.Tick) error {
	slice.Lock()
	if slice.len == 0 || len(slice.holdings) == 0 {
		slice.Unlock()
		return utils.ErrEmptySlice
	}
	for _, holding := range slice.holdings {
		holding.UpdateMetrics(tick)
	}
	slice.Unlock()
	return nil
}

// ApplyDelta ... TODO
func (slice *HoldingSlice) ApplyDelta(amt utils.Amount) error {
	slice.Lock()
	newVolume := slice.totalVolume + amt
	if newVolume < 0 {
		slice.Unlock()
		return ErrNegativeVolume
	}
	slice.totalVolume = newVolume
	slice.Unlock()
	return nil
}

// Push adds position to position slice,
// updates total Volume of all positions in slice.
func (slice *HoldingSlice) Push(pos *holding.Holding) error {
	slice.holdings = append(slice.holdings, pos)
	return slice.ApplyDelta(pos.Volume)
}

// Pop removes element from position slice.
// If fifo is passed as costmethod, the position at index 0 will be popped.
// Otherwise if lifo is passed as costmethod, the position at the last index will be popped.
func (slice *HoldingSlice) Pop(costMethod utils.CostMethod) (holding.Holding, error) {
	var pos holding.Holding

	slice.Lock()

	switch costMethod {
	case utils.Fifo:
		pos, slice.holdings = *slice.holdings[0], slice.holdings[1:]
	case utils.Lifo:
		pos, slice.holdings = *slice.holdings[slice.len], slice.holdings[:slice.len-1]
	}
	slice.len--

	slice.Unlock()
	return pos, nil
}

// Peek returns the element that would have been Pop-ed from a holding slice.
func (slice *HoldingSlice) Peek(costMethod utils.CostMethod) (*holding.Holding, error) {
	var holding *holding.Holding

	if slice.len == 0 || len(slice.holdings) == 0 {
		return nil, utils.ErrEmptySlice
	}

	slice.RLock()
	switch costMethod {
	case utils.Fifo:
		holding = slice.holdings[0]
	case utils.Lifo:
		holding = slice.holdings[slice.len]
	}
	slice.RUnlock()
	return holding, nil
}

// LookupCache acts as read-write cache to indicate index positions of holdings in a slice.
type LookupCache struct {
	items     map[string]int16
	mu        sync.RWMutex
	openSlots []int16
	last      int16
}

// NewLookupCache creates a new LookupCache.
func NewLookupCache() *LookupCache {
	return &LookupCache{
		items:     make(map[string]int16),
		openSlots: make([]int16, 0),
		last:      -1,
	}
}

// Delete removes cached key-value pair, allocates index to openSlots.
func Delete(l *LookupCache, key string) {
	l.mu.Lock()
	value := Get(l, key)
	delete(l.items, key)

	if value != -1 {
		l.openSlots = append(l.openSlots, value)
	}
	l.mu.Unlock()
}

// Get queries for the index (value) loaded, returns -1 if key does not exist.
func Get(l *LookupCache, key string) int16 {
	l.mu.RLock()
	value, ok := l.items[key]
	l.mu.RUnlock()

	if !ok {
		return -1
	}
	return value
}

// Put assigns a value to a new key.
// If any pre-allocated slots are available, it will be assigned that slot.
// If not, it will get a new, unassigned value.
func Put(l *LookupCache, key string) (err error) {
	var value int16

	if Get(l, key) != -1 {
		return errors.New("key already exists in LookupCache")
	}

	l.mu.Lock()
	len := len(l.openSlots)

	if len > 0 {
		value, l.openSlots = l.openSlots[0], l.openSlots[1:]
	} else {
		l.last++
		value = l.last
	}
	l.items[key] = value
	l.mu.Unlock()

	return nil
}

// LinkedHoldingList is a collection of holding elements,
// as well as aggregate metrics on the collection of holdings.
type LinkedHoldingList struct {
	*instrument.Asset
	mu   sync.RWMutex
	head *LinkedHoldingNode
	tail *LinkedHoldingNode
}

// NewLinkedHoldingList instantiates a new struct of type LinkedHoldingList.
func NewLinkedHoldingList(h *holding.Holding, t *instrument.Tick) *LinkedHoldingList {
	l := &LinkedHoldingList{
		Asset: instrument.NewAsset(t.Ticker, t.Bid, t.Ask, h.Volume, t.Timestamp),
		head:  new(LinkedHoldingNode),
		tail:  newLinkedHoldingNode(h),
	}
	l.head.next = l.tail
	l.tail.prev = l.head

	return l
}

// Push inserts a new element
func (l *LinkedHoldingList) Push(node *LinkedHoldingNode) {
	var last *LinkedHoldingNode

	switch l.head.next == nil {
	case true:
		last = l.head
	case false:
		last = l.tail
	}
	last.next = node
	node.prev = last
	l.tail = last
}

// Pop returns last element in linkedList.
// Returns nil if no elements in list besides head and tail.
func (l *LinkedHoldingList) Pop() *LinkedHoldingNode {
	last := l.tail
	if last == l.head {
		return nil // cannot pop head
	}
	l.tail = last.prev
	l.tail.next = nil
	return last
}

// PopFront ...TODO
func (l *LinkedHoldingList) PopFront() {

}

// Peek ...TODO
func (l *LinkedHoldingList) Peek() {

}

// PeekFront ...TODO
func (l *LinkedHoldingList) PeekFront() {

}

// LinkedHoldingNode is the linked list node implementation of a holding struct.
type LinkedHoldingNode struct {
	*holding.Holding
	next *LinkedHoldingNode
	prev *LinkedHoldingNode
}

func newLinkedHoldingNode(h *holding.Holding) *LinkedHoldingNode {
	return &LinkedHoldingNode{
		Holding: h,
		next:    nil,
		prev:    nil,
	}
}
