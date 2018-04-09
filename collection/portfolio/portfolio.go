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
	sync.RWMutex
	Active map[string]*collection.HoldingSlice
	cash   utils.Amount
}

// New creates a new instance of a Portfolio struct.
func New() *Portfolio {
	port := Portfolio{
		Active: make(map[string]*collection.HoldingSlice),
		cash:   0,
	}
	return &port
}

// ApplyDelta ... TODO
func (port *Portfolio) ApplyDelta(amt utils.Amount) error {
	port.Lock()
	newBalance := port.cash + amt
	if newBalance < 0 {
		return ErrNegativeCash
	}
	port.cash = newBalance
	port.Unlock()
	return nil
}

// UpdateMetrics updates holding metric data based off of tick information.
func (port *Portfolio) UpdateMetrics(tick instrument.Tick) error {
	_, exists := port.Active[tick.Ticker]
	if !exists {
		return utils.ErrEmptySlice
	}

	return port.Active[tick.Ticker].UpdateMetrics(tick)
}

// AddHolding adds a new holding to the dedicated Active holdingSlice if it exists.
// Returns error if holdingSlice with same ticker does not exist in map.
func (port *Portfolio) AddHolding(newHolding *holding.Holding, deltaCash utils.Amount) error {
	port.Lock()
	if _, ok := port.Active[newHolding.Ticker]; !ok {
		// log.Printf("Created new slice for %s", newHolding.Ticker)
		port.Active[newHolding.Ticker] = collection.NewHoldingSlice()
	}
	err := port.Active[newHolding.Ticker].AddNew(newHolding)
	if err != nil {
		port.Unlock()
		return err
	}
	port.Unlock()
	return port.ApplyDelta(deltaCash)
}
