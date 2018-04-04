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

		cash:        cashAmt,
		deltaChan:   make(chan Amount),
		balanceChan: make(chan Amount),
		errChan:     make(chan error),
	}
	go port.mux()
	return &port
}

// Portfolio struct refer to the aggregation of positions traded by a broker.
type Portfolio struct {
	sync.RWMutex

	active map[string]*HoldingSlice

	cash        Amount
	deltaChan   chan Amount
	balanceChan chan Amount
	errChan     chan error
}

func (port *Portfolio) mux() {
	var deltaCash Amount

	for {
		select {
		case deltaCash = <-port.deltaChan:
			port.errChan <- port.applyDelta(deltaCash)

		case port.balanceChan <- port.cash:
		}
	}
}

// Cash returns most recent version of portfolio balance
func (port *Portfolio) Cash() Amount {
	return <-port.balanceChan
}

// UpdateCash updates portfolio cash balance
func (port *Portfolio) UpdateCash(amt Amount) error {
	port.deltaChan <- amt
	return <-port.errChan
}

func (port *Portfolio) applyDelta(amt Amount) error {
	newBalance := port.cash + amt
	if newBalance < 0 {
		return ErrNegativeCash
	}
	port.cash = newBalance
	return nil
}

// UpdateMetrics updates holding metric data based off of tick information.
func (port *Portfolio) UpdateMetrics(tick Tick) error {
	_, exists := port.active[tick.Ticker]
	if !exists {
		return ErrEmptySlice
	}

	err := port.active[tick.Ticker].UpdateMetrics(tick)
	return err
}

// AddNew adds a new holdingSlice to the active map.
func (port *Portfolio) AddNew(pos *Position, deltaCash Amount) error {
	port.RLock()
	_, exists := port.active[pos.Ticker]
	port.RUnlock()

	if exists {
		return ErrSliceExists
	}
	port.active[pos.Ticker] = NewHoldingSlice()
	slice := port.active[pos.Ticker]

	if err := slice.addNew(pos); err != nil {
		return err
	}
	return port.UpdateCash(deltaCash)
}

// AddHolding adds a new holding to the dedicated active holdingSlice if it exists.
// Returns error if holdingSlice with same ticker does not exist in map.
func (port *Portfolio) AddHolding(newHolding *Position, deltaCash Amount) error {
	port.RLock()
	_, exists := port.active[newHolding.Ticker]
	port.RUnlock()

	if !exists {
		return ErrNoSliceExists
	}
	port.Lock()
	if err := port.active[newHolding.Ticker].addNew(newHolding); err != nil {
		port.Unlock()
		return err
	}
	port.Unlock()
	return port.UpdateCash(deltaCash)
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

// UpdateMetrics passes tick to appropriate Security in securities map.
// Returns error if security not found in map.
func (index *Index) UpdateMetrics(tick Tick) error {
	index.tickChan <- tick
	return <-index.errChan
}

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

// NewHoldingSlice returns a new holding slice and starts its channel multiplexer.
func NewHoldingSlice() *HoldingSlice {
	holdingSlice := &HoldingSlice{
		len:         0,
		totalVolume: 0,
		holdings:    make([]*Position, 0),

		deltaLenChan: make(chan int),

		pushChan:   make(chan *Position),
		popChan:    make(chan CostMethod),
		poppedChan: make(chan *Position),
		peekChan:   make(chan *Position),

		tickChan: make(chan Tick),

		deltaVolumeChan: make(chan Amount),
		volumeChan:      make(chan Amount),
		errChan:         make(chan error),
	}
	go holdingSlice.mux()
	return holdingSlice
}

// HoldingSlice is a slice that holds pointer values to Position type variables
// TODO(HoldingSlice) update description
type HoldingSlice struct {
	sync.RWMutex
	len          int
	deltaLenChan chan int

	holdings   []*Position
	pushChan   chan *Position
	peekChan   chan *Position
	popChan    chan CostMethod
	poppedChan chan *Position

	tickChan chan Tick

	totalVolume     Amount
	volumeChan      chan Amount
	deltaVolumeChan chan Amount
	errChan         chan error
}

func (slice *HoldingSlice) mux() {
	var deltaVolume Amount
	var pushedPos *Position

	for {
		select {

		case deltaVolume = <-slice.deltaVolumeChan:
			slice.errChan <- slice.applyDelta(deltaVolume)

		case pushedPos = <-slice.pushChan:
			slice.errChan <- slice.push(pushedPos)

		case costMethod := <-slice.popChan:
			popped, err := slice.pop(costMethod)
			slice.poppedChan <- &popped
			slice.errChan <- err

		case newLen := <-slice.deltaLenChan:
			slice.errChan <- slice.updateLen(newLen)

		case tick := <-slice.tickChan:
			slice.errChan <- slice.updateMetrics(tick)

		}
	}
}

// AddNew adds a new holding to the holdings slice.
func (slice *HoldingSlice) addNew(newHolding *Position) error {
	slice.Lock()
	slice.holdings = append(slice.holdings, newHolding)
	slice.len++
	slice.Unlock()

	slice.UpdateVolume(newHolding.Volume)
	return nil
}

// UpdateMetrics update holdings data based on most recent tick data.
func (slice *HoldingSlice) UpdateMetrics(tick Tick) error {
	slice.tickChan <- tick
	return <-slice.errChan
}

// UpdateMetrics update holdings data based on most recent tick data.
func (slice *HoldingSlice) updateMetrics(tick Tick) error {
	slice.Lock()
	if slice.len == 0 || len(slice.holdings) == 0 {
		return ErrEmptySlice
	}
	for _, holding := range slice.holdings {
		holding.UpdateMetrics(tick)
	}
	slice.Unlock()
	return nil
}

// Volume returns the position slice's total volume.
func (slice *HoldingSlice) Volume() Amount {
	return <-slice.volumeChan
}

// UpdateVolume updates a position slice's total volume.
func (slice *HoldingSlice) UpdateVolume(amt Amount) error {
	slice.deltaVolumeChan <- amt
	return <-slice.errChan
}

func (slice *HoldingSlice) applyDelta(amt Amount) error {
	newVolume := slice.totalVolume + amt
	if newVolume < 0 {
		log.Println("woops")
		return ErrNegativeVolume
	}
	slice.totalVolume = newVolume
	return nil
}

// UpdateLen updates total number of position pointers held in positions
func (slice *HoldingSlice) UpdateLen(len int) error {
	slice.deltaLenChan <- len
	return <-slice.errChan
}
func (slice *HoldingSlice) updateLen(len int) error {
	slice.len = len
	return nil
}

// Push adds position to position slice,
// updates total Volume of all positions in slice.
func (slice *HoldingSlice) Push(pos *Position) error {
	slice.pushChan <- pos
	return <-slice.errChan
}

func (slice *HoldingSlice) push(pos *Position) error {
	slice.holdings = append(slice.holdings, pos)
	slice.deltaVolumeChan <- pos.Volume
	return <-slice.errChan
}

// Pop removes element from position slice.
// If fifo is passed as costmethod, the position at index 0 will be popped.
// Otherwise if lifo is passed as costmethod, the position at the last index will be popped.
func (slice *HoldingSlice) Pop(costMethod CostMethod) (pos Position, err error) {
	slice.popChan <- costMethod
	pos = *<-slice.poppedChan
	err = <-slice.errChan
	return
}

func (slice *HoldingSlice) pop(costMethod CostMethod) (Position, error) {
	var pos Position

	slice.Lock()
	if slice.len == 0 {
		slice.Unlock()
		return Position{}, ErrEmptySlice
	}
	switch costMethod {
	case fifo:
		pos = *slice.holdings[0]
		slice.holdings = slice.holdings[1:]
	case lifo:
		pos = *slice.holdings[slice.len]

		slice.holdings = slice.holdings[:slice.len-1]
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
