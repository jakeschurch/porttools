package porttools

import "errors"

// CostMethod regards the type of accounting mangagement rule
// is implemented for selling securities
type CostMethod int

const (
	FIFO CostMethod = iota
	LIFO
)

// Portfolio struct holds instances of finacial stock Positions.
type Portfolio struct {
	ActivePositions PositionSlice
	ClosedPositions PositionSlice
	Orders          []*Order
	Cash            float64
	Benchmark       *Security
}

// Transact conducts agreement between Position and Order within a portfolio
func (portfolio *Portfolio) Transact(order *Order) (err error) {
	// TODO: Money management function to make sure enough cash is on hand

	// Add order to the Portfolio's Orders slice
	portfolio.Orders = append(portfolio.Orders, order)

	switch order.TransactionT {

	case Sell:
		for order.Volume > 0 {
			positionToSell := portfolio.ActivePositions.LookupToSell(order.Ticker, FIFO)
			portfolio.Sell(positionToSell, order)
		}
	case Buy:
		portfolio.Buy(order)
	}

	return nil
}

// Buy is a function that creates a new position based off of an Order
// and appends it to a Portfolio's ActivePositions posSlice.
func (portfolio *Portfolio) Buy(order *Order) *Position {
	portfolio.Cash -= order.Volume * order.Price

	positionToBuy := Position{Ticker: order.Ticker, Volume: order.Volume,
		BoughtAt: datedMetric{Amount: order.Price, Date: order.Date}}

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
	} // QUESTION: Throw error?
}

// PositionSlice is a posSlice that holds pointer values to Position type variables
type PositionSlice []*Position

// Remove is a function to remove *position values from a posSlice in-place.
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

func (portfolio *Portfolio) updatePositions(tick *Tick) {
	for _, pos := range portfolio.ActivePositions {
		if pos.Ticker == tick.Ticker {
			pos.update(*tick)
		}
	}
}

func (posSlice PositionSlice) LookupToSell(ticker string, costMethod CostMethod) *Position {
	var positionFound *Position
Loop:
	for _, pos := range posSlice {

		if pos.Ticker == ticker {

			if positionFound.BoughtAt.Date.IsZero() {
				positionFound = pos
				continue Loop
			}

			switch costMethod {

			case FIFO:
				if pos.BoughtAt.Date.Before(positionFound.BoughtAt.Date) {
					positionFound = pos
					continue Loop
				}

			case LIFO:
				if pos.BoughtAt.Date.After(positionFound.BoughtAt.Date) {
					positionFound = pos
					continue Loop
				}
			}
		}
	}
	return positionFound
}

// MarketPortfolio struct holds instances of finacial stock securities.
type MarketPortfolio struct {
	Instruments map[string]*Security
}

// AddtoPortfolio adds Position instrument to MarketPortfolio instance.
func (p *MarketPortfolio) AddtoPortfolio(s *Security) (err error) {
	if _, exists := p.Instruments[s.Ticker]; !exists {
		p.Instruments[s.Ticker] = s
		return nil
	}
	return errors.New("Security already exists in Instruments map")
}
