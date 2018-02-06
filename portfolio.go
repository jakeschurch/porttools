package porttools

import "errors"

// TODO: rename/factor AddToPortfolio functions

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

	switch order.TransactionT {
	case Sell:
		portfolio.Sell(position, order)

	case Buy:
		position.Volume += order.Volume
		portfolio.ActivePositions = append(portfolio.ActivePositions, position)
	}

	// Add order to the Portfolio's Orders slice
	portfolio.Orders = append(portfolio.Orders, order)

	return nil
}

func (portfolio *Portfolio) Buy(order *Order) (*Position, error) {


// Sell is a function that removes a Position's volume, as well as create
// a new closed position. Updates a portfolio's cash balance.
func (portfolio *Portfolio) Sell(position *Position, order *Order) (err error) {

	if activeVolume := position.Volume - order.Volume; activeVolume >= 0 {
		soldPosition := *position

		soldPosition.sellShares(order)

		// Update active volume for position
		position.Volume = activeVolume

		// TODO: update cash balace

		portfolio.ClosedPositions = append(portfolio.ClosedPositions, &soldPosition)

		if activeVolume == 0 {
			portfolio.ActivePositions.Remove(position)
		}
	}
	return nil
}

// AddtoPortfolio adds Position instrument to MarketPortfolio instance.
func (p *MarketPortfolio) AddtoPortfolio(s *Security) (err error) {
	if _, exists := p.Instruments[s.Ticker]; !exists {
		p.Instruments[s.Ticker] = s
		return nil
	}
	return errors.New("Security already exists in Instruments map")

}
