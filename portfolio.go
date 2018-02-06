package porttools

import "errors"

// PositionSlice TODO
type PositionSlice []*Position

// Remove is a function to remove *position values from a slice in-place.
func (slice PositionSlice) Remove(toRemove *Position) PositionSlice {
	j := 0
	for _, val := range slice {
		if val != toRemove {
			slice[j] = val
			j++
		}
	}
	return slice[:j]
}

// MarketPortfolio struct holds instances of finacial stock securities.
type MarketPortfolio struct {
	Instruments map[string]*Security
}

// Portfolio struct holds instances of finacial stock Positions.
type Portfolio struct {
	ActivePositions PositionSlice
	ClosedPositions PositionSlice
	Orders          []*Order
	Cash            float64
	Benchmark       *Security
}

// Transact conducts agreement between Position and Order within a portfolio
// TODO(Transact): Init/find active positions
func (portfolio *Portfolio) Transact(position *Position, order *Order) (err error) {
	// TODO: Money management function to make sure enough cash is on hand

	// Add order to the Portfolio's Orders slice
	portfolio.Orders = append(portfolio.Orders, order)

	switch order.TransactionT {
	case Sell:
		portfolio.Sell(position, order)

	case Buy:
		portfolio.Buy(order)
	}

	return nil
}

// Buy is a function that creates a new position based off of an Order
// and appends it to a Portfolio's ActivePositions slice.
func (portfolio *Portfolio) Buy(order *Order) *Position {
	portfolio.Cash -= order.Volume * order.Price

	position := Position{Ticker: order.Ticker, Volume: order.Volume,
		PriceBought: order.Price, DateBought: order.Date}

	portfolio.ActivePositions = append(portfolio.ActivePositions, &position)

	return &position
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

// AddtoPortfolio adds Position instrument to MarketPortfolio instance.
func (p *MarketPortfolio) AddtoPortfolio(s *Security) (err error) {
	if _, exists := p.Instruments[s.Ticker]; !exists {
		p.Instruments[s.Ticker] = s
		return nil
	}
	return errors.New("Security already exists in Instruments map")

}
