package porttools

import (
	"errors"
	"log"
	"sync"
)

// NewPortfolio creates a new instance of a Portfolio struct.
func NewPortfolio(cashAmt Amount) *Portfolio {
	return &Portfolio{
		Active: make(map[string]*PositionSlice),
		Cash:   cashAmt,
	}
}

// Portfolio struct refer to the aggregation of positions traded by a broker.
type Portfolio struct {
	sync.RWMutex
	Cash   Amount                    `json:"cash"`
	Active map[string]*PositionSlice `json:"active"`
}

func (port *Portfolio) updatePositions(tick Tick) {
	for _, pos := range port.Active[tick.Ticker].positions {
		if pos.LastBid.Date.Before(tick.Timestamp) {
			break // REVIEW return instead of break?
		}
		pos.updateMetrics(&tick)
	}
}

// TODO REVIEW: how to add easy accessibility for populating index without having
// to scan two maps? (being Instruments map & toIgnore map)
func newIndex() *Index {
	return &Index{
		Instruments: make(map[string]*Security),
	}
}

// Index struct allow for the use of a benchmark to compare financial performance,
// Index could refer to one Security or many.
type Index struct {
	Instruments map[string]*Security
	sync.Mutex
}

func (index *Index) updateSecurity(tick Tick) (ok bool) {
	index.Lock()
	if security, exists := index.Instruments[tick.Ticker]; !exists {
		index.Instruments[tick.Ticker] = NewSecurity(tick)
	} else {
		security.updateMetrics(tick)
	}
	index.Unlock()
	return true
}

func newPositionSlice() *PositionSlice {
	return &PositionSlice{
		len:         0,
		positions:   make([]*Position, 0),
		totalVolume: 0,
	}
}

// PositionSlice is a slice that holds pointer values to Position type variables
type PositionSlice struct {
	sync.Mutex  // COMBAK: may need this later...or not.
	positions   []*Position
	len         int
	totalVolume Amount
}

// Push adds position to position slice,
// updates total Volume of all positions in slice.
func (slice *PositionSlice) Push(pos *Position) {
	slice.Lock()
	slice.totalVolume += pos.Volume
	slice.positions = append(slice.positions, pos)
	slice.Unlock()
	return
}

// Pop removes element from position slice.
// If fifo is passed as costmethod, the position at index 0 will be popped.
// Otherwise if lifo is passed as costmethod, the position at the last index will be popped.
func (slice *PositionSlice) Pop(costMethod CostMethod) (pos *Position, err error) {
	if slice.len == 0 {
		log.Fatalf("%s position slice has underflowed.\n", pos.Ticker)
		return nil, errors.New("Buffer underflow")
	}
	slice.Lock()

	switch costMethod {
	case fifo:
		pos = slice.positions[0]
		slice.positions = slice.positions[1:]
	case lifo:
		pos = slice.positions[slice.len]
		slice.positions = slice.positions[:slice.len-1]
	}
	slice.len--
	slice.totalVolume -= pos.Volume
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
