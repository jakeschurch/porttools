package porttools

import (
	"sync"

	"github.com/jakeschurch/porttools/collection"
	"github.com/jakeschurch/porttools/collection/portfolio"
	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/order"
	"github.com/jakeschurch/porttools/utils"
)

// OMS acts as an `Order Management System` to test trading signals and fill orders.
type OMS struct {
	mu   sync.RWMutex
	open *collection.HoldingList
	cash utils.Amount
}

// NewOMS inits a new OMS type.
func NewOMS() *OMS {
	oms := &OMS{
		open: collection.NewHoldingList(),
		cash: 0,
	}
	return oms
}

func (oms *OMS) Insert(o *order.Order) error {
	var dxCash utils.Amount

	switch o.Buy {
	case true:
		dxCash = -o.Ask * o.Volume(0)

	case false:
		dxCash = o.Bid * o.Volume(0)
	}
	oms.updateCash(dxCash)
	return oms.open.Insert(o)
}

func (oms *OMS) Execute(node *collection.LinkedNode) {

}

func (oms *OMS) Query(q *instrument.Quote) {

	var orderList *collection.LinkedList
	var openOrder *collection.LinkedNode
	var entryOrder, exitOrder *order.Order
	var err error

	switch entryOrder, _ := strategy.CheckEntryLogic(q); entryOrder != nil {
	case true:
		oms.Insert(entryOrder)
	case false: // do nothing if entry logic is not met.
	}

	orderList, _ = oms.open.Get(q.Ticker())
	for openOrder = orderList.PeekFront(); openOrder != nil; openOrder = openOrder.Next() {

		// TEMP: for now, do nothing with exitOrder
		exitOrder, err = strategy.CheckExitLogic(openOrder.GetUnderlying().(order.Order))
		switch err != nil {
		case false:

		case true: // do nothing if invalid exit logic.

		}
	}

}

func (oms *OMS) QueryBuyOrder() {

}

func (oms *OMS) QuerySellOrder() {

}

func (oms *OMS) updateCash(dxCash utils.Amount) {
	oms.mu.Lock()
	oms.cash += dxCash
	oms.mu.Unlock()

	return
}

func (oms *OMS) addOrder(newOrder *order.Order) (utils.Amount, error) {
	oms.Lock()
	oms.openOrders = append(oms.openOrders, newOrder)
	oms.Unlock()
	return -(newOrder.Volume * newOrder.Ask), nil
}

// TransactSell will sell an order and update a holding slice to reflect the changes.
func (oms *OMS) TransactSell(order *order.Order, costMethod utils.CostMethod, port *portfolio.Portfolio) (utils.Amount, []instrument.Holding, error) {
	var closedHoldings []instrument.Holding
	var transactionAmount utils.Amount
	var sellVolume utils.Amount
	var pos *instrument.Holding
	var err error

	// loop over slice until order has been completely transacted
	for order.Volume > 0 {

		pos, err = port.Active[order.Ticker].Peek(costMethod)
		if err != nil {
			// log.Println(err)
			return transactionAmount, closedHoldings, err
		}
		switch pos.Volume >= order.Volume {
		case true:
			sellVolume = order.Volume
		case false:
			sellVolume = pos.Volume
		}

		// update cash and remove sold volume from Active holding
		port.Active[order.Ticker].ApplyDelta(-sellVolume)
		pos.Volume -= sellVolume

		// create new closed position
		bid := &utils.DatedMetric{Amount: order.Bid, Date: order.Datetime}
		ask := &utils.DatedMetric{Amount: order.Ask, Date: order.Datetime}
		newClosedPosition := instrument.Holding{
			Ticker:   pos.Ticker,
			Volume:   sellVolume,
			NumTicks: pos.NumTicks,
			LastBid:  bid, LastAsk: ask,
			AvgBid: pos.AvgBid, AvgAsk: pos.AvgAsk,
			MaxBid: pos.MaxBid, MaxAsk: pos.MaxAsk,
			MinBid: pos.MinBid, MinAsk: pos.MinAsk,
			BuyPrice: pos.BuyPrice, SellPrice: ask,
		}
		// add new closed position to closedHoldings slice
		closedHoldings = append(closedHoldings, newClosedPosition)

		// Update total of transaction utils.Amount
		transactionAmount += (ask.Amount * sellVolume)

		if pos.Volume == 0 {
			port.Active[order.Ticker].Pop(costMethod)
		}
		order.Volume -= sellVolume
	}
	oms.closeOutOrder(order)

	return transactionAmount, closedHoldings, nil
}

func (oms *OMS) closeOutOrder(closedOrder *order.Order) {
	orders := make([]*order.Order, 0)
	oms.Lock()
	for _, order := range oms.openOrders {
		if order != closedOrder {
			orders = append(orders, order)
		}
	}
	oms.openOrders = orders
	oms.Unlock()
}
