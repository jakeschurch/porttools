package porttools

import "time"

// PortfolioLog conducts performance analysis.
type PortfolioLog struct {
	Closed    positionSlice
	orders    Queue
	benchmark *Index
}

// func report() {
//
// }

/* TODO:
- max-drawdown
- % profitable
- total num trades
- winning/losing trades
- trading period length
*/

// CostMethod regards the type of accounting management rule
// is implemented for selling securities.
type CostMethod int

const (
	lifo CostMethod = iota - 1
	fifo
)

// Order struct hold information referring to the
// details of an execution of a financial asset transaction.
type Order struct {
	// it's either buy or sell
	Buy      bool
	Status   OrderStatus
	Logic    TradeLogic
	Ticker   string
	Price    Amount
	Volume   Amount
	Datetime time.Time
}

// OrderStatus variables refer to a status of an order's execution.
type OrderStatus int

const (
	open OrderStatus = iota // 0
	completed
	canceled
	expired // 3
)

// TradeLogic is used to identify when the order should be executed.
type TradeLogic int

const (
	market TradeLogic = iota // 0
	limit
	stopLimit
	stopLoss
	dayTrade // 4
)

// Transact conducts agreement between Position and Order within a portfolio.
func (port *Portfolio) Transact(order *Order, costMethod CostMethod) (err error) {
	// TODO: Money management function to make sure enough cash is on hand & securities

	// Add order to the Portfolio's Orders slice
	// NOTE: may want to store in order book out of struct via channel instead
	port.Orders = append(port.Orders, order)

	switch order.Buy {
	case false:
		switch costMethod {
		case fifo:
			port.sell(*order, costMethod)
		case lifo:
			port.sell(*order, costMethod)
		}
	case true:
		port.buy(*order)
	}
	return nil
}

// buy is a function that creates a new Position based off of an Order
// and appends it to a Portfolio's position's slice.
func (port *Portfolio) buy(order Order) *Position {
	port.Cash -= order.Volume * order.Price
	posBought := &Position{
		Ticker: order.Ticker, Volume: order.Volume,
		BuyPrice: datedMetric{order.Price, order.Datetime},
	}
	port.Active[order.Ticker].Push(posBought)

	// = append(port.Active[order.Ticker], posBought)

	return posBought
}

// sell is a function that removes a Position's volume, as well as create
// a new closed position. Updates a port's cash balance.
func (port *Portfolio) sell(order Order, costMethod CostMethod) {
	ticker := order.Ticker

	// TODO: check to see if we have full amount to sell. return bool/err
	port.Cash += order.Volume * order.Price
	for order.Volume > 0 {
		posToSell := port.Active[ticker].Peek(costMethod)

		var closedPos *Position
		if posToSell.Volume >= order.Volume {
			closedPos = posToSell.sellShares(order, order.Volume)
		} else {
			closedPos = posToSell.sellShares(order, posToSell.Volume)
		}
		port.Closed[order.Ticker].Push(closedPos)

		if posToSell.Volume == 0 {
			port.Closed[ticker].Pop(costMethod)
		}
	}
}

func (pos *Position) sellShares(order Order, amount Amount) *Position {

	soldPos := func() *Position {
		posSold := *pos
		posSold.Volume = amount
		posSold.SellPrice = datedMetric{order.Price, order.Datetime}
		return &posSold
	}()
	// Update active volume for pos
	pos.Volume = pos.Volume - amount

	return soldPos
}
