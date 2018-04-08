package porttools

import (
	"sync"

	"github.com/jakeschurch/porttools/collection/portfolio"
	"github.com/jakeschurch/porttools/instrument/holding"
	"github.com/jakeschurch/porttools/trading/order"
	"github.com/jakeschurch/porttools/utils"
)

// OMS acts as an `Order Management System` to test trading signals and fill orders.
type OMS struct {
	sync.RWMutex
	openOrders []*order.Order
}

// NewOMS inits a new OMS type.
func NewOMS() *OMS {
	oms := &OMS{
		openOrders: make([]*order.Order, 0),
	}
	return oms
}

func (oms *OMS) addOrder(newOrder *order.Order) (utils.Amount, error) {
	oms.Lock()
	oms.openOrders = append(oms.openOrders, newOrder)
	oms.Unlock()
	return -(newOrder.Volume * newOrder.Ask), nil
}

func (oms *OMS) existsInOrders(ticker string) ([]*order.Order, error) {
	matchedOrders := make([]*order.Order, 0)
	oms.RLock()
	for _, order := range oms.openOrders {
		if order.Ticker == ticker {
			matchedOrders = append(matchedOrders, order)
		}
	}
	oms.RUnlock()
	if len(matchedOrders) == 0 {
		return nil, utils.ErrEmptySlice
	}
	return matchedOrders, nil
}

// TransactSell will sell an order and update a holding slice to reflect the changes.
func (oms *OMS) TransactSell(order *order.Order, costMethod utils.CostMethod, port *portfolio.Portfolio) (utils.Amount, []holding.Holding, error) {
	var closedHoldings []holding.Holding
	var transactionAmount utils.Amount
	var sellVolume utils.Amount
	var pos *holding.Holding
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
		newClosedPosition := holding.Holding{
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
