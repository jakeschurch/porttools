package porttools

import "sync"

func newPositionSlice() *positionSlice {
	return &positionSlice{len: 0, positions: make([]*Position, 0)}
}

// PositionSlice is a slice that holds pointer values to Position type variables
type positionSlice struct {
	len       int
	positions []*Position
	// tickChan chan Tick
	// ticker String
	// totalPositions
}

// Push adds position to position slice,
// updates total Volume of all positions in slice.
func (slice *PositionSlice) Push(pos *Position) {
	slice.len++
	if slice.len-1 == 0 {
		slice.positions[0] = pos
		return
	}
	slice.positions[slice.len] = pos
	return
}

// Pop removes element from position slice.
// If fifo is passed as costmethod, the position at index 0 will be popped.
// Otherwise if lifo is passed as costmethod, the position at the last index will be popped.
func (slice *PositionSlice) Pop(costMethod CostMethod) (pos *Position, err error) {
	if slice.len == 0 {
		return nil, errors.New("Buffer underflow")
	}
	switch costMethod {
	case fifo:
		pos = slice.positions[0]
		slice.positions = slice.positions[0:]
	case lifo:
		pos = slice.positions[slice.len]
		slice.positions = slice.positions[:slice.len]
	}
	slice.len--
	return
}

// Peek
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

// Portfolio struct refer to the aggregation of positions traded by a broker.
type Portfolio struct {
	Active    map[string]*positionSlice `json:"active"`
	Closed    map[string]*positionSlice `json:"closed"`
	Orders    []*Order                  `json:"orders"` // NOTE: may not need this
	Cash      Amount                    `json:"cash"`
	Benchmark *Index                    `json:"benchmark"`
	mutex     sync.Mutex
	// IDEA: max/min equity as datedmetrics
}

// TODO Look into concurrent access of struct pointers
func (port *Portfolio) updatePosition(tick Tick) {
	for _, pos := range port.Active[tick.Ticker].positions {
		pos.updateMetrics(tick)
	}
}

// Index structs allow for the use of a benchmark to compare financial performance,
// Index could refer to one Security or many.
type Index struct {
	Instruments map[string]*Security
}

func (index *Index) updateMetrics(tick *Tick) (ok bool) {
	if security, exists := index.Instruments[tick.Ticker]; !exists {
		index.Instruments[tick.Ticker] = NewSecurity(*tick)
	} else {
		security.updateMetrics(*tick)
	}
	return true
}
