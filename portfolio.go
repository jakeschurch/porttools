package porttools

// CostMethod regards the type of accounting mangagement rule
// is implemented for selling securities.
type CostMethod int

const (
	lifo CostMethod = iota - 1
	fifo
)

// Portfolio structs refer to the aggregation of positions traded by a broker.
type Portfolio struct {
	ActivePositions map[string][]*Position
	ClosedPositions map[string][]*Position
	Orders          []*Order
	Cash            Currency
	Benchmark       *Index
}

// Transact conducts agreement between Position and Order within a portfolio.
func (port *Portfolio) Transact(order *Order, costMethod CostMethod) (err error) {
	// TODO: Money management function to make sure enough cash is on hand & securities

	// Add order to the Portfolio's Orders slice
	port.Orders = append(port.Orders, order)

	switch order.Buy {
	case false:

		switch costMethod {
		case fifo:
			port.Sell(order, costMethod)
		case lifo:
			port.Sell(order, CostMethod(len(port.ActivePositions[order.Ticker])-1))
		}
	case true:
		port.Buy(order)
	}
	return nil
}

// Buy is a function that creates a new Position based off of an Order
// and appends it to a Portfolio's ActivePositions posSlice.
func (port *Portfolio) Buy(order *Order) *Position {
	port.Cash -= order.Volume * order.Price

	posBought := &Position{
		Ticker: order.Ticker, Volume: order.Volume,
		BuyPrice: datedMetric{order.Price, order.Datetime}}

	port.ActivePositions[order.Ticker] = append(port.ActivePositions[order.Ticker], posBought)

	return posBought
}

// Sell is a function that removes a Position's volume, as well as create
// a new closed posToSell. Updates a port's cash balance.
func (port *Portfolio) Sell(order *Order, costMethod CostMethod) {
	ticker := order.Ticker
	for order.Volume > 0 {

		posToSell := port.ActivePositions[ticker][costMethod]
		port.Cash += order.Volume * order.Price

		var closedPos *Position
		if posToSell.Volume >= order.Volume {
			closedPos = posToSell.sellShares(order, order.Volume)
		} else {
			closedPos = posToSell.sellShares(order, posToSell.Volume)
		}
		port.ClosedPositions[order.Ticker] = append(port.ClosedPositions[order.Ticker], closedPos)

		if posToSell.Volume == 0 {
			port.RemoveFromActive(posToSell)
		}
	}
}

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

// TODO Look into concurrent access of struct pointers
func (port *Portfolio) updatePosition(tick *Tick) {
	for _, pos := range port.ActivePositions[tick.Ticker] {
		if pos.Ticker == tick.Ticker {
			pos.update(*tick)
		}
	}
}

// Index structs allow for the use of a benchmark to compare financial performance,
// Index could refer to one Security or many.
type Index struct {
	Instruments map[string]*Security
}

func (index *Index) update(tick *Tick) (ok bool) {
	if security, exists := index.Instruments[tick.Ticker]; !exists {
		index.Instruments[tick.Ticker] = NewSecurity(*tick)
	} else {
		security.update(*tick)
	}
	return true
}
