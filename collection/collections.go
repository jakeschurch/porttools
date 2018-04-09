package collection

import (
	"errors"
	"sync"

	"github.com/jakeschurch/porttools/instrument/security"

	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/instrument/holding"
	"github.com/jakeschurch/porttools/utils"
)

// const (
// 	SecurityType = "Security"
// 	HoldingType  = "Holding"
// )

var (
	// ErrKeyExists indicates that an index value already exists in the lookup cache
	ErrKeyExists = errors.New("key already exists in LookupCache")

	// ErrSliceExists indicates that a slice already exists
	ErrSliceExists = errors.New("slice with ticker exists")

	ErrListNotEmpty = errors.New("linked holding list is still populated with at least one element")

	// ErrNoListExists indicates that a slice already exists
	ErrNoListExists = errors.New("list with ticker does not exist")

	// ErrNegativeVolume indicates negative position volume balance
	ErrNegativeVolume = errors.New("position volume is less than 0")
)

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
func Delete(l *LookupCache, key string) int16 {
	l.mu.Lock()
	value := Get(l, key)
	delete(l.items, key)

	if value != -1 {
		l.openSlots = append(l.openSlots, value)
	}
	l.mu.Unlock()
	return value
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

// Put assigns a value to a new key, and returns the value of the newly-assigned index..
// If any pre-allocated slots are available, it will be assigned that slot.
// If not, it will get a new, unassigned value.
func Put(l *LookupCache, key string) (value int16, err error) {

	if value = Get(l, key); value != -1 {
		return value, ErrKeyExists
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

	return value, nil
}

// LinkedList is a collection of holding elements,
// as well as aggregate metrics on the collection of holdings.
type LinkedList struct {
	*instrument.Asset
	mu       sync.RWMutex
	head     *LinkedNode
	tail     *LinkedNode
	nodeType string
}

// NewLinkedList instantiates a new struct of type LinkedList.
func NewLinkedList(h *holding.Holding, t instrument.Tick) *LinkedList {
	l := &LinkedList{
		Asset: instrument.NewAsset(t.Ticker, t.Bid, t.Ask, h.Volume(0), t.Timestamp),
		head:  new(LinkedNode),
		tail:  NewLinkedNode(h),
	}
	l.head.next = l.tail
	l.tail.prev = l.head

	return l
}

// Push inserts a new element
func (l *LinkedList) Push(f instrument.Financial, t instrument.Tick) {
	var last *LinkedNode

	// Update LinkedList Aggregates
	l.Update(t)
	l.Volume(f.Volume(0))
	node := NewLinkedNode(f)

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
func (l *LinkedList) Pop(costMethod utils.CostMethod) *LinkedNode {
	if costMethod == utils.Lifo {
		return l.pop()
	}
	return l.PopFront()
}

// Pop returns last element in linkedList.
// Returns nil if no elements in list besides head and tail.
func (l *LinkedList) pop() *LinkedNode {
	last := l.tail
	if last == l.head {
		return nil // cannot pop head
	}
	l.tail = last.prev
	l.tail.next = nil

	l.Volume(-last.Volume(0))
	return last
}

// PopFront ...TODO
func (l *LinkedList) PopFront() *LinkedNode {
	var first *LinkedNode

	if first = l.head.next; first == nil {
		return nil
	}

	l.head.next = first.next
	first.next.prev = l.head

	l.mu.Lock()
	l.Volume(-first.Volume(0))
	l.mu.Unlock()

	return first
}

// Peek ...TODO
func (l *LinkedList) Peek() {
}

// PeekFront ...TODO
func (l *LinkedList) PeekFront() *LinkedNode {
	return l.head.next
}

// LinkedNode is the linked list node implementation of a holding struct.
type LinkedNode struct {
	instrument.Financial
	next *LinkedNode
	prev *LinkedNode
}

func (node *LinkedNode) GetUnderlying() instrument.Financial {
	switch node.Financial.(type) {
	case holding.Holding:
		return holding.Holding(node.Financial.(holding.Holding))

	case security.Security:
		return security.Security(node.Financial.(security.Security))
	}
	return nil
}

func (node *LinkedNode) Next() *LinkedNode {
	return node.next
}

func NewLinkedNode(f instrument.Financial) *LinkedNode {
	return &LinkedNode{
		Financial: f,
		next:      nil,
		prev:      nil,
	}
}

// HoldingList is an implementation of a holding collection.
type HoldingList struct {
	cache *LookupCache
	mu    sync.RWMutex
	list  []*LinkedList
	len   int16
}

// NewHoldingList returns a new struct of type HoldingList.
func NewHoldingList() *HoldingList {
	return &HoldingList{
		cache: NewLookupCache(),
		list:  make([]*LinkedList, 0),
		len:   0,
	}
}

func (l *HoldingList) Update(t instrument.Tick) error {
	var index int16

	if index = Get(l.cache, t.Ticker); index == -1 {
		return ErrNoListExists
	}
	return l.list[index].Update(t)
}

// Get method for type HoldingList returns a LinkedList and error types.
func (l *HoldingList) Get(key string) (*LinkedList, error) {
	var index int16
	var linkedList *LinkedList

	if index = Get(l.cache, key); index != -1 {
		l.mu.RLock()
		linkedList = l.list[index]
		l.mu.RUnlock()
		return linkedList, nil
	}
	return nil, ErrNoListExists
}

func (l *HoldingList) GetByIndex(index int16) *LinkedList {
	l.mu.RLock()
	linkedList := l.list[index]
	l.mu.RUnlock()

	return linkedList
}

// Insert adds a new node to a HoldingList's linked list.
func (l *HoldingList) Insert(h *holding.Holding, t instrument.Tick) (err error) {
	var new bool
	var index int16

	if index, err = Put(l.cache, h.Ticker); err != ErrKeyExists {
		new = true
	}

	// see if we can place new holding in open slot
	// ... or if we have to allocate new space.
	l.mu.Lock()
	if index > l.len {
		l.list = append(make([]*LinkedList, (index+1)*2), l.list...)
		l.len = (index + 1) * 2
	} else {
		l.len++
	}
	l.mu.Unlock()

	// Check to see if we need to allocate a new Linked list
	// ... or if we can just push new node.
	switch new {
	case true:
		l.list[index] = NewLinkedList(h, t)
	case false:
		l.list[index].Push(NewLinkedNode(h), t)
	}
	return nil
}

// Delete ... TODO
func (l *HoldingList) Delete(key string) error {
	var index int16
	var linkedHoldings *LinkedList

	if index = Get(l.cache, key); index == -1 {
		return ErrNoListExists
	}

	l.mu.Lock()
	if linkedHoldings = l.list[index]; linkedHoldings.tail != nil {
		return ErrListNotEmpty
	}
	l.list[index] = nil
	l.mu.Unlock()

	Delete(l.cache, key)

	return nil
}

// Items returns items map of a lookup cache.
func (l *HoldingList) Items() map[string]int16 {
	return l.cache.items
}
