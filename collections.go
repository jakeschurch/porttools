package porttools

// PositionSlice is a slice that holds pointer values to Position type variables
type positionSlice []*Position

// RemoveFromActive delete's a Portfolio's active position.
func (port *Portfolio) RemoveFromActive(toRemove *Position) {
	j := 0
	pSlice := port.ActivePositions[toRemove.Ticker]
	for _, val := range pSlice {
		if val != toRemove {
			pSlice[j] = val
			j++
		}
	}
}

// Portfolio structs refer to the aggregation of positions traded by a broker.
type Portfolio struct {
	ActivePositions map[string][]*Position
	ClosedPositions map[string][]*Position
	Orders          []*Order
	Cash            Amount
	Benchmark       *Index
}

// TODO Look into concurrent access of struct pointers
func (port *Portfolio) updatePosition(tick *Tick) {
	for _, pos := range port.ActivePositions[tick.Ticker] {
		if pos.Ticker == tick.Ticker {
			pos.updateMetrics(*tick)
		}
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
