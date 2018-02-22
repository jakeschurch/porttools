package porttools

// CostMethod regards the type of accounting mangagement rule
// is implemented for selling securities
type CostMethod int

const (
	fifo CostMethod = iota
	lifo
)

// Portfolio structs refer to the aggregation of positions traded by a broker.
type Portfolio struct {
	ActivePositions PositionSlice
	ClosedPositions PositionSlice
	Orders          []*Order
	Cash            float64
	Benchmark       *Index
}

// Transact conducts agreement between Position and Order within a portfolio
func (portfolio *Portfolio) Transact(order *Order) (err error) {
	// TODO: Money management function to make sure enough cash is on hand

	// Add order to the Portfolio's Orders slice
	portfolio.Orders = append(portfolio.Orders, order)

	switch order.TransactionType {

	case Sell:
		for order.Volume > 0 {
			positionToSell := portfolio.ActivePositions.lookupToSell(order.Ticker, fifo)
			portfolio.Sell(positionToSell, order)
		}
	case Buy:
		portfolio.Buy(order)
	}

	return nil
}

// Buy is a function that creates a new Position based off of an Order
// and appends it to a Portfolio's ActivePositions posSlice.
func (portfolio *Portfolio) Buy(order *Order) *Position {
	portfolio.Cash -= order.Volume * order.Price

	positionToBuy := Position{Ticker: order.Ticker, Volume: order.Volume,
		BuyPrice: datedMetric{Amount: order.Price, Date: order.Datetime}}

	portfolio.ActivePositions = append(portfolio.ActivePositions, &positionToBuy)

	return &positionToBuy
}

// Sell is a function that removes a Position's volume, as well as create
// a new closed position. Updates a portfolio's cash balance.
func (portfolio *Portfolio) Sell(position *Position, order *Order) {

	if activeVolume := position.Volume - order.Volume; activeVolume >= 0 {
		soldPosition := *position
		soldPosition.sellShares(order)

		// Update active volume for position
		position.Volume = activeVolume

		portfolio.ClosedPositions = append(portfolio.ClosedPositions, &soldPosition)
		portfolio.Cash += order.Volume * order.Price

		if activeVolume == 0 {
			portfolio.ActivePositions.Remove(position)
		}
	}
}

// PositionSlice is a slice that holds pointer values to Position type variables
type PositionSlice []*Position

// Remove is a function to remove *Position values from a posSlice in-place.
func (posSlice PositionSlice) Remove(toRemove *Position) PositionSlice {
	j := 0
	for _, val := range posSlice {
		if val != toRemove {
			posSlice[j] = val
			j++
		}
	}
	return posSlice[:j]
}

// TODO Look into concurrent access of struct pointers
func (portfolio *Portfolio) updatePosition(tick *Tick) {
	for _, pos := range portfolio.ActivePositions {
		if pos.Ticker == tick.Ticker {
			pos.update(*tick)
		}
	}
}

func (posSlice PositionSlice) lookupToSell(ticker string, costMethod CostMethod) *Position {
	var positionFound *Position
Loop:
	for _, pos := range posSlice {

		if pos.Ticker == ticker {
			if positionFound.BuyPrice.Date.IsZero() {
				positionFound = pos
				continue Loop
			}

			switch costMethod {
			case fifo:
				if pos.BuyPrice.Date.Before(positionFound.BuyPrice.Date) {
					positionFound = pos
					continue Loop
				}
			case lifo:
				if pos.BuyPrice.Date.After(positionFound.BuyPrice.Date) {
					positionFound = pos
					continue Loop
				}
			}
		}
	}
	return positionFound
}

// Index structs allow for the use of a benchmark to compare financial performance,
// Index could refer to one Security or many.
type Index struct {
	Instruments map[string]*Security
}

func (index *Index) update(tick *Tick) (ok bool) {
	ok = true
	if security, exists := index.Instruments[tick.Ticker]; !exists {
		index.Instruments[tick.Ticker] = NewSecurity(*tick)
	} else {
		security.update(*tick)
	}
	return ok
}
