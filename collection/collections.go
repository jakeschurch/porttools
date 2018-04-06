package collection

import (
	"errors"
	"log"
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
		log.Println(slice.holdings)
		pos, slice.holdings = *slice.holdings[slice.len], slice.holdings[:slice.len-1]
		log.Println(slice.holdings)
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
