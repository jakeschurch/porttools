package collection

import (
	"errors"
	"sync"

	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/order"
	"github.com/jakeschurch/porttools/utils"
)

var (
	// ErrNodeNotFound indicates that a node could not be found in a linked list.
	ErrNodeNotFound = errors.New("node not foudn in linked list")

	ErrEmptyList = errors.New("list is empty; please delete")

	// ErrKeyExists indicates that an index value already exists in the lookup cache
	ErrKeyExists = errors.New("key already exists in LookupCache")

	// ErrSliceExists indicates that a slice already exists
	ErrSliceExists = errors.New("slice with ticker exists")

	ErrListNotEmpty = errors.New("linked holding list is still populated with at least one element")

	// ErrNoListExists indicates that a slice already exists
	ErrNoListExists = errors.New("list with ticker does not exist")
)

// ------------------------------------------------------------------

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

// ------------------------------------------------------------------

// LinkedNode is the linked list node implementation of a holding struct.
type LinkedNode struct {
	instrument.Financial
	next *LinkedNode
	prev *LinkedNode
}

func NewLinkedNode(f instrument.Financial) *LinkedNode {
	return &LinkedNode{
		Financial: f,
		next:      nil,
		prev:      nil,
	}
}

func (node *LinkedNode) GetUnderlying() instrument.Financial {

	switch node.Financial.(type) {
	case instrument.Holding:
		return instrument.Holding(node.Financial.(instrument.Holding))

	case instrument.Security:
		return instrument.Security(node.Financial.(instrument.Security))

	case order.Order:
		return order.Order(node.Financial.(order.Order))

	default:
		return nil
	}
}

func (node *LinkedNode) Next() *LinkedNode {
	return node.next
}

// ------------------------------------------------------------------

// LinkedList is a collection of holding elements,
// as well as aggregate metrics on the collection of holdings.
type LinkedList struct {
	*instrument.Asset
	mu   sync.RWMutex
	head *LinkedNode
	tail *LinkedNode
}

// NewLinkedList instantiates a new struct of type LinkedList.
func NewLinkedList(f instrument.Financial) *LinkedList {
	var asset instrument.Asset

	switch f.(type) {
	case instrument.Asset:
		asset = f.(instrument.Asset)

	case instrument.Security:
		asset = f.(instrument.Security).GetUnderlying().(instrument.Asset)
	}

	l := &LinkedList{
		Asset: &asset,
		head:  new(LinkedNode),
		tail:  NewLinkedNode(f),
	}
	l.head.next = l.tail
	l.tail.prev = l.head

	return l
}

// Push inserts a new element
func (l *LinkedList) Push(f instrument.Financial) {
	var last *LinkedNode

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

func (l *LinkedList) PopToSecurity(c utils.CostMethod) *instrument.Security {
	popped := l.Pop(c)
	security := &instrument.Security{
		Asset: *l.Asset,
	}
	security.Volume(-security.Volume(0) + popped.Volume(0))
	security.Nticks = l.Nticks - popped.Financial.(instrument.Holding).Nticks

	return security
}

func (l *LinkedList) PeekToSecurity(newVolume utils.Amount, c utils.CostMethod) *instrument.Security {
	peeked := l.Peek(c)
	peeked.Volume(-newVolume)

	security := &instrument.Security{
		Asset: *l.Asset,
	}
	security.Volume(-security.Volume(0) + newVolume)
	security.Nticks = l.Nticks - peeked.Financial.(instrument.Holding).Nticks

	return security
}

// Pop returns last element in linkedList.
// Returns nil if no elements in list besides head and tail.
func (l *LinkedList) Pop(c utils.CostMethod) *LinkedNode {
	if c == utils.Lifo {
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

func (l *LinkedList) Peek(c utils.CostMethod) *LinkedNode {
	if c == utils.Lifo {
		return l.peek()
	}
	return l.PeekFront()
}

// Peek ...TODO
func (l *LinkedList) peek() *LinkedNode {
	return l.tail
}

// PeekFront ...TODO
func (l *LinkedList) PeekFront() *LinkedNode {
	return l.head.next
}

func (l *LinkedList) remove(node *LinkedNode) error {
	var next = l.head.next

	switch node {
	case l.tail:
		l.pop()
		return nil

	case next:
		l.PopFront()
		return nil

	default:
		next = next.next
	}

	for next != nil {
		if next == node { // delete node reference
			next.prev.next = next.next
			next.next.prev = next.prev

			l.Volume(-next.Volume(0))
			//  linkedList
			if l.head.next == nil {
				return ErrEmptyList
			}
			return nil
		}
		next = next.next
	}
	return ErrNodeNotFound
}

// ------------------------------------------------------------------

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

func (l *HoldingList) Update(q instrument.Quote) error {
	var index int16

	if index = Get(l.cache, q.Ticker()); index == -1 {
		return ErrNoListExists
	}
	return l.list[index].Update(q)
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

func (l *HoldingList) RemoveNode(node *LinkedNode) error {
	var list *LinkedList
	var err error

	if list, err = l.Get(node.Ticker()); err != nil {
		if err == ErrNoListExists {
			return nil
		}
		return err
	}
	if err = list.remove(node); err != nil {
		if err == ErrEmptyList {
			return l.Delete(node.Ticker())
		}
		return err
	}
	return nil
}

func (l *HoldingList) GetByIndex(index int16) *LinkedList {
	l.mu.RLock()
	linkedList := l.list[index]
	l.mu.RUnlock()

	return linkedList
}

// Insert adds a new node to a HoldingList's linked list.
func (l *HoldingList) Insert(f instrument.Financial) (err error) {
	var new bool
	var index int16

	if index, err = Put(l.cache, f.Ticker()); err != ErrKeyExists {
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
		l.list[index] = NewLinkedList(f)
	case false:
		l.list[index].Push(NewLinkedNode(f))
	}
	return nil
}

func (l *HoldingList) InsertUpdate(f instrument.Financial, q instrument.Quote) (err error) {
	var new bool
	var index int16

	if index, err = Put(l.cache, f.Ticker()); err != ErrKeyExists {
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
		l.list[index] = NewLinkedList(f)
	case false:
		l.list[index].Push(NewLinkedNode(f))
	}
	l.GetByIndex(index).Update(q)
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
