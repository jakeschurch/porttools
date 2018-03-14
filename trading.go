package porttools

import (
	"errors"
	"time"
)

// func report() {
//
// }

//
// - max-drawdown
// - % profitable
// - total num trades
// - winning/losing trades
// - trading period length

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
func (port *Portfolio) Transact(order *Order, costMethod CostMethod) error {

	switch order.Buy {
	case true:
		ok := func() bool { // check to see if order can be fulfilled.
			port.RLock()
			ok := (order.Volume * order.Price) <= port.Cash
			port.RUnlock()
			return ok
		}()
		if !ok { // if not ok, cancel order and return error.
			order.Status = canceled
			port.Orders = append(port.Orders, order)
			return errors.New("Not enough cash to fulfil order")
		}

		// Create new Position and add it to according position slice.
		posBought := &Position{
			Ticker: order.Ticker, Volume: order.Volume,
			BuyPrice: datedMetric{order.Price, order.Datetime},
		}
		port.Active[order.Ticker].Push(posBought)
		port.Active[order.Ticker].totalAmt += posBought.Volume // Update position slice volume.

	case false: // sell
		// Check to see if order can be fulfilled, if not, cancel order and return error.
		if port.Active[order.Ticker].totalAmt < order.Volume {
			order.Status = canceled
			port.Orders = append(port.Orders, order)
			return errors.New("Not enough volume to satisfy order")
		}
		switch costMethod {
		case fifo:
			port.sell(*order, costMethod)
		case lifo:
			port.sell(*order, costMethod)
		}
	}
	order.Status = completed
	// Add order to the Portfolio's Orders slice
	// NOTE: may want to store in order book out of struct via channel instead
	port.Orders = append(port.Orders, order)

	return nil
}

// sell is a function that removes a Position's volume, as well as create
// a new closed position. Updates a port's cash balance.
func (port *Portfolio) sell(order Order, costMethod CostMethod) (err error) {
	// Update Cash Amount.
	port.Lock()
	port.Cash += order.Volume * order.Price
	port.Unlock()

	for order.Volume > 0 {
		var posToSell *Position
		if posToSell = port.Active[order.Ticker].Peek(costMethod); posToSell == nil {
			err = errors.New("No position to sell")
			return
		}

		var sellAmt Amount
		if posToSell.Volume >= order.Volume {
			sellAmt = order.Volume
		} else {
			sellAmt = posToSell.Volume
		}
		posToSell.Volume -= sellAmt

		closedPos := *posToSell
		closedPos.Volume = sellAmt
		closedPos.SellPrice = datedMetric{order.Price, order.Datetime}

		port.Lock()
		port.Closed[order.Ticker].Push(&closedPos)

		if posToSell.Volume == 0 {
			port.Closed[order.Ticker].Pop(costMethod)
		}
		port.Unlock()
	}
	return nil
}
