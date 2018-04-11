package collection

import (
	"errors"
	"log"
	"sync"

	ins "github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/utils"
)

// Declare Errors
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

	if len(l.openSlots) > 0 {
		value, l.openSlots = l.openSlots[0], l.openSlots[1:]
	} else {
		l.last++
		value = l.last
	}
	l.items[key] = value
	l.mu.Unlock()

	return value, nil
}

func AddOpenSlots(l *LookupCache, slots ...int16) {
	l.mu.Lock()
	for n := range slots {
		l.openSlots = append(l.openSlots, slots[n])
		l.last++
	}
	l.mu.Unlock()
}

// ------------------------------------------------------------------

// LinkedNode is the linked list node implementation of a holding struct.
type LinkedNode struct {
	Data ins.Instrument
	next *LinkedNode
	prev *LinkedNode
}

// NewLinkedNode creates a Linked List Element.
func NewLinkedNode(data ins.Instrument) *LinkedNode {
	return &LinkedNode{
		Data: data,
		next: nil,
		prev: nil,
	}
}

// Next returns pointer to next-point element, or nil.
func (node *LinkedNode) Next() *LinkedNode {
	return node.next
}

// ------------------------------------------------------------------

// LinkedList is a collection of holding elements,
// as well as aggregate metrics on the collection of holdings.
type LinkedList struct {
	*ins.AssetSumm
	mu   sync.RWMutex
	head *LinkedNode
	tail *LinkedNode
}

// NewLinkedList instantiates a new struct of type LinkedList.
func NewLinkedList(a ins.AssetSumm, i ins.Instrument) *LinkedList {

	l := &LinkedList{
		AssetSumm: &a,
		head:      new(LinkedNode),
		tail:      NewLinkedNode(i),
	}
	l.head.next = l.tail
	l.tail.prev = l.head

	l.TotalVolume += ins.ExtractQuote(i).Volume

	return l
}

// Push inserts a new element
func (l *LinkedList) Push(i ins.Instrument) {
	var last, node *LinkedNode

	l.mu.Lock()
	l.TotalVolume += ins.ExtractQuote(i).Volume
	l.mu.Unlock()

	node = NewLinkedNode(i)

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

func (l *LinkedList) PopToSecurity(c utils.CostMethod, o ins.Order) (*ins.Security, error) {
	var popped *LinkedNode

	if popped = l.Pop(c); popped == nil {
		return nil, nil
	}

	return ins.SellOff(popped.Data, o, l.AssetSumm)
}

func (l *LinkedList) PeekToSecurity(c utils.CostMethod, o ins.Order) (*ins.Security, error) {
	var peeked *LinkedNode

	if peeked = l.Peek(c); peeked == nil {
		return nil, nil
	}

	return ins.SellOff(peeked.Data, o, l.AssetSumm)
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

	l.TotalVolume -= ins.ExtractQuote(last.Data).Volume
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
	l.TotalVolume -= first.Data.(ins.Quote).Volume
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

			l.TotalVolume -= ins.ExtractQuote(next.Data).Volume
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
		list:  make([]*LinkedList, 1),
		len:   0,
	}
}

func (l *HoldingList) Update(q ins.Quote) error {
	var index int16

	if index = Get(l.cache, q.Ticker); index == -1 {
		return ErrNoListExists
	}
	if l.list[index] != nil {
		l.list[index].AssetSumm.Update(q)
		return nil
	}
	return nil
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

	if list, err = l.Get(ins.ExtractQuote(node.Data).Ticker); err != nil {
		if err == ErrNoListExists {
			return nil
		}
		return err
	}
	if err = list.remove(node); err != nil {
		if err == ErrEmptyList {
			return l.Delete(node.Data.(ins.Quote).Ticker)
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
func (l *HoldingList) Insert(i ins.Instrument) error {
	var new bool
	var quote = *ins.ExtractQuote(i)
	var index, err = Put(l.cache, quote.Ticker)
	if err != ErrKeyExists {
		new = true
	}
	// see if we can place new holding in open slot
	// ... or if we have to allocate new space.
	l.mu.Lock()
	if index >= l.len || l.len == 0 {
		oldLen := l.len
		log.Print(oldLen)
		l.len = (index + 1) * 2
		l.list = append(make([]*LinkedList, l.len+1), l.list...)

		var tempList = make([]int16, 0)
		var n int16
		for n = oldLen + 1; n < l.len; n++ {
			tempList = append(tempList, n)
		}
		AddOpenSlots(l.cache, tempList...)
	}
	l.mu.Unlock()

	// Check to see if we need to allocate a new Linked list
	// ... or if we can just push new node.
	switch new {
	case true:
		quote := *ins.ExtractQuote(i)
		assetSumm := *ins.NewAssetSumm(quote)
		l.list[index] = NewLinkedList(assetSumm, i)
	case false:
		l.list[index].Push(i)
	}
	return nil
}

func (l *HoldingList) InsertUpdate(i ins.Instrument, q ins.Quote) (err error) {
	var new bool
	var index int16

	if index, err = Put(l.cache, ins.ExtractQuote(i).Ticker); err != ErrKeyExists {
		new = true
	}

	// see if we can place new holding in open slot
	// ... or if we have to allocate new space.
	l.mu.Lock()
	if index > l.len || l.len == 0 {
		l.list = append(make([]*LinkedList, (index+1)*2), l.list...)
		l.len = (index + 1) * 2
	}
	l.mu.Unlock()

	// Check to see if we need to allocate a new Linked list
	// ... or if we can just push new node.
	switch new {
	case true:
		quote := *ins.ExtractQuote(i)
		assetSumm := *ins.NewAssetSumm(quote)
		l.list[index] = NewLinkedList(assetSumm, i)
	case false:
		l.list[index].Push(i)
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
