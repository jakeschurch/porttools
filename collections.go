package porttools

import (
	"errors"
	"log"
	"sync"
)

func newPositionSlice() *PositionSlice {
	return &PositionSlice{
		len:         0,
		positions:   make([]*Position, 0),
		totalVolume: 0,
	}
}

// PositionSlice is a slice that holds pointer values to Position type variables
type PositionSlice struct {
	positions   []*Position
	len         int
	totalVolume Amount
	*sync.Mutex // COMBAK: may need this later...or not.
}

// Push adds position to position slice,
// updates total Volume of all positions in slice.
func (slice *PositionSlice) Push(pos *Position) {
	slice.Lock()
	slice.len++

	if slice.len-1 == 0 {
		slice.positions[0] = pos
		slice.Unlock()
		return
	}
	slice.positions[slice.len] = pos
	slice.Unlock()
	return
}

// Pop removes element from position slice.
// If fifo is passed as costmethod, the position at index 0 will be popped.
// Otherwise if lifo is passed as costmethod, the position at the last index will be popped.
func (slice *PositionSlice) Pop(costMethod CostMethod) (pos *Position, err error) {
	if slice.len == 0 {
		log.Printf("%s position slice has underflowed.\n", pos.Ticker)
		return nil, errors.New("Buffer underflow")
	}
	slice.Lock()
	switch costMethod {
	case fifo:
		pos = slice.positions[0]
		slice.positions = slice.positions[0:]
	case lifo:
		pos = slice.positions[slice.len]
		slice.positions = slice.positions[:slice.len]
	}
	slice.len--
	slice.Unlock()
	return
}

// Peek returns the element that would have been Pop-ed from the position slice.
func (slice *PositionSlice) Peek(costMethod CostMethod) (pos *Position) {
	if slice.len == 0 {
		return nil
	}
	switch costMethod {
	case fifo:
		pos = slice.positions[0]
	case lifo:
		pos = slice.positions[slice.len]
	}
	return
}

// NewPortfolio creates a new instance of a Portfolio struct.
func NewPortfolio(cashAmt Amount) *Portfolio {
	return &Portfolio{
		Active: make(map[string]*PositionSlice),
		Closed: make(map[string]*PositionSlice),
		Cash:   cashAmt,
		Orders: make([]*Order, 0),
	}
}

// Portfolio struct refer to the aggregation of positions traded by a broker.
type Portfolio struct {
	Active map[string]*PositionSlice `json:"active"`
	Closed map[string]*PositionSlice `json:"closed"`
	Orders []*Order                  `json:"orders"`
	// NOTE: may not be the `best` idea to store Orders in Portfolio struct.
	Cash Amount `json:"cash"`
	*sync.RWMutex
}

func (port *Portfolio) updatePositions(tick *Tick) {
	for _, pos := range port.Active[tick.Ticker].positions {
		if pos.LastBid.Date.Before(tick.Timestamp) {
			break
			// REVIEW return instead of break?
		}
		go pos.updateMetrics(tick)
	}
}

// Index structs allow for the use of a benchmark to compare financial performance,
// Index could refer to one Security or many.
type Index struct {
	Instruments map[string]*Security
	*sync.Mutex
}

func (index *Index) updateSecurity(tick *Tick) (ok bool) {
	if security, exists := index.Instruments[tick.Ticker]; !exists {
		index.Lock()
		index.Instruments[tick.Ticker] = NewSecurity(*tick)
		index.Unlock()
	} else {
		security.updateMetrics(*tick)
	}
	return true
}
