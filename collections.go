package porttools

import (
	"errors"
	"log"
	"sync"
)

var (
	// ErrNegativeCash indicates negative cash balance
	ErrNegativeCash = errors.New("Insufficient Funds")

	// ErrNegativeVolume indicates negative position volume balance
	ErrNegativeVolume = errors.New("Position Volume is less than 0")

	// ErrEmptySlice indicates a slice with 0 elements
	ErrEmptySlice = errors.New("Slice has 0 elements")

	// ErrSliceExists indicates that a slice already exists
	ErrSliceExists = errors.New("Slice with ticker exists")

	// ErrNoSliceExists indicates that a slice already exists
	ErrNoSliceExists = errors.New("Slice with ticker does not exist")

	// ErrSecurityExists indicates that a security struct has already been allocated in an index
	ErrSecurityExists = errors.New("Security exists in index")

	// ErrNoSecurityExists indicates that a security struct has not been allocated in an index's securities map
	ErrNoSecurityExists = errors.New("Security does not exist in Securities map")
)

// NewPortfolio creates a new instance of a Portfolio struct.
func NewPortfolio(cashAmt Amount) *Portfolio {
	port := Portfolio{
		active: make(map[string]*HoldingSlice),
		cash:   cashAmt,
	}
	return &port
}

// Portfolio struct refer to the aggregation of positions traded by a broker.
type Portfolio struct {
	sync.RWMutex
	active map[string]*HoldingSlice
	cash   Amount
}

func (port *Portfolio) applyDelta(amt Amount) error {
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
func (port *Portfolio) UpdateMetrics(tick Tick) error {
	_, exists := port.active[tick.Ticker]
	if !exists {
		return ErrEmptySlice
	}

	err := port.active[tick.Ticker].updateMetrics(tick)
	return err
}

// AddHolding adds a new holding to the dedicated active holdingSlice if it exists.
// Returns error if holdingSlice with same ticker does not exist in map.
func (port *Portfolio) AddHolding(newHolding *Position, deltaCash Amount) error {
	port.Lock()
	if _, ok := port.active[newHolding.Ticker]; !ok {
		log.Printf("Created new slice for %s", newHolding.Ticker)
		port.active[newHolding.Ticker] = NewHoldingSlice()
	}
	err := port.active[newHolding.Ticker].addNew(newHolding)
	if err != nil {
		port.Unlock()
		return err
	}
	port.Unlock()
	return port.applyDelta(deltaCash)
}

// NewIndex returns a new Index type; typically used for benchmarking a portfolio.
func NewIndex() *Index {
	index := Index{
		Securities: make(map[string]*Security),
		tickChan:   make(chan Tick),
		errChan:    make(chan error),
	}
	go index.mux()
	return &index
}

// Index struct allow for the use of a benchmark to compare financial performance,
// Index could refer to one Security or many.
type Index struct {
	sync.RWMutex
	Securities map[string]*Security
	tickChan   chan Tick
	errChan    chan error
}

func (index *Index) mux() {
	var tick Tick
	for {
		select {
		case tick = <-index.tickChan:
			index.errChan <- index.updateMetrics(tick)
		}
	}
}

// AddNew adds a new security to an Index's Securities map.
func (index *Index) AddNew(tick Tick) error {
	index.Lock()
	if _, exists := index.Securities[tick.Ticker]; exists {
		return ErrSecurityExists
	}
	index.Securities[tick.Ticker] = NewSecurity(tick)
	index.Unlock()

	return nil
}

// UpdateMetrics ...
func (index *Index) UpdateMetrics(tick Tick) error {
	index.tickChan <- tick
	return <-index.errChan
}

// UpdateMetrics passes tick to appropriate Security in securities map.
// Returns error if security not found in map.
func (index *Index) updateMetrics(tick Tick) error {
	index.RLock()
	security, exists := index.Securities[tick.Ticker]
	index.RUnlock()
	if !exists {
		return ErrNoSecurityExists
	}
	security.updateMetrics(tick)
	return nil
}

// NewHoldingSlice returns a new holding slice.
func NewHoldingSlice() *HoldingSlice {
	holdingSlice := &HoldingSlice{
		len:         0,
		totalVolume: 0,
		holdings:    make([]*Position, 0),
	}
	return holdingSlice
}

// HoldingSlice is a slice that holds pointer values to Position type variables
type HoldingSlice struct {
	sync.RWMutex
	len         int
	holdings    []*Position
	totalVolume Amount
}

// AddNew adds a new holding to the holdings slice.
func (slice *HoldingSlice) addNew(newHolding *Position) error {
	// slice.Lock()
	slice.holdings = append(slice.holdings, newHolding)
	slice.len++
	// slice.Unlock()

	return slice.applyDelta(newHolding.Volume)
}

// UpdateMetrics update holdings data based on most recent tick data.
func (slice *HoldingSlice) updateMetrics(tick Tick) error {
	slice.Lock()
	if slice.len == 0 || len(slice.holdings) == 0 {
		slice.Unlock()
		return ErrEmptySlice
	}
	for _, holding := range slice.holdings {
		holding.UpdateMetrics(tick)
	}
	slice.Unlock()
	return nil
}

func (slice *HoldingSlice) applyDelta(amt Amount) error {
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
func (slice *HoldingSlice) push(pos *Position) error {
	slice.holdings = append(slice.holdings, pos)
	return slice.applyDelta(pos.Volume)
}

// Pop removes element from position slice.
// If fifo is passed as costmethod, the position at index 0 will be popped.
// Otherwise if lifo is passed as costmethod, the position at the last index will be popped.
func (slice *HoldingSlice) pop(costMethod CostMethod) (Position, error) {
	var pos Position

	slice.Lock()
	// if slice.len == 0 {
	// 	slice.Unlock()
	// 	return Position{}, ErrEmptySlice
	// }
	switch costMethod {
	case fifo:
		pos, slice.holdings = *slice.holdings[0], slice.holdings[1:]
	case lifo:
		log.Println(slice.holdings)
		pos, slice.holdings = *slice.holdings[slice.len], slice.holdings[:slice.len-1]
		log.Println(slice.holdings)
	}
	slice.len--

	slice.Unlock()
	return pos, nil
}

// Peek returns the element that would have been Pop-ed from a holding slice.
func (slice *HoldingSlice) Peek(costMethod CostMethod) (*Position, error) {
	var holding *Position

	if slice.len == 0 || len(slice.holdings) == 0 {
		return nil, ErrEmptySlice
	}

	slice.RLock()
	switch costMethod {
	case fifo:
		holding = slice.holdings[0]
	case lifo:
		holding = slice.holdings[slice.len]
	}
	slice.RUnlock()
	return holding, nil
}
