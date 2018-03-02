package porttools

import "time"

// CostMethod regards the type of accounting mangagement rule
// is implemented for selling securities.
type CostMethod int

const (
	lifo CostMethod = iota - 1
	fifo
)

// IDEA orderbook || TransactionLog

// Order structs hold information referring to the details of the execution of a financial asset transaction.
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
	day // 4
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
			port.sell(*order, CostMethod(len(port.ActivePositions[order.Ticker])-1))
		}
	case true:
		port.buy(*order)
	}
	return nil
}

// buy is a function that creates a new Position based off of an Order
// and appends it to a Portfolio's ActivePositions posSlice.
// IDEA instead of appending to new slice just reindex
func (port *Portfolio) buy(order Order) *Position {
	port.Cash -= order.Volume * order.Price

	posBought := &Position{
		Ticker: order.Ticker, Volume: order.Volume,
		BuyPrice: datedMetric{order.Price, order.Datetime},
	}

	port.ActivePositions[order.Ticker] = append(port.ActivePositions[order.Ticker], posBought)

	return posBought
}

// sell is a function that removes a Position's volume, as well as create
// a new closed posToSell. Updates a port's cash balance.
func (port *Portfolio) sell(order Order, costMethod CostMethod) {
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
