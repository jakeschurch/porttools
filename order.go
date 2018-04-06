package porttools

import (
	"log"
	"sync"
	"time"
)

// OMS acts as an `Order Management System` to test trading signals and fill orders.
type OMS struct {
	sync.RWMutex
	openOrders []*Order
}

// NewOMS inits a new OMS type.
func NewOMS() *OMS {
	oms := &OMS{
		openOrders: make([]*Order, 0),
	}
	return oms
}

func (oms *OMS) addOrder(newOrder *Order) (Amount, error) {
	oms.Lock()
	oms.openOrders = append(oms.openOrders, newOrder)
	oms.Unlock()
	return -(newOrder.Volume * newOrder.Ask), nil
}

func (oms *OMS) existsInOrders(ticker string) ([]*Order, error) {
	matchedOrders := make([]*Order, 0)
	oms.RLock()
	for _, order := range oms.openOrders {
		if order.Ticker == ticker {
			matchedOrders = append(matchedOrders, order)
		}
	}
	oms.RUnlock()
	if len(matchedOrders) == 0 {
		return nil, ErrEmptySlice
	}
	return matchedOrders, nil
}

// TransactSell will sell an order and update a holding slice to reflect the changes.
func (oms *OMS) TransactSell(order *Order, costMethod CostMethod, port *Portfolio) (Amount, []Position, error) {
	var closedHoldings []Position
	var transactionAmount Amount
	var sellVolume Amount
	var holding *Position
	var err error

	// holdingSlice.RLock()
	// if holdingSlice == nil {
	// 	holdingSlice.RUnlock()
	// 	return transactionAmount, []*Position{}, deleteSlice, ErrEmptySlice
	// }
	// if holdingSlice.len == 0 {
	// 	holdingSlice.RUnlock()
	// 	return transactionAmount, []*Position{}, deleteSlice, ErrEmptySlice
	// }
	// holdingSlice.RUnlock()

	// loop over slice until order has been completely transacted
	for order.Volume > 0 {

		holding, err = port.active[order.Ticker].Peek(costMethod)
		if err != nil {
			log.Println(err)
			return transactionAmount, closedHoldings, err
		}
		switch holding.Volume >= order.Volume {
		case true:
			sellVolume = order.Volume
		case false:
			sellVolume = holding.Volume
		}

		// update cash and remove sold volume from active holding
		port.active[order.Ticker].applyDelta(-sellVolume)
		holding.Volume -= sellVolume

		// create new closed position
		bid := &datedMetric{Amount: order.Bid, Date: order.Datetime}
		ask := &datedMetric{Amount: order.Ask, Date: order.Datetime}
		newClosedPosition := Position{
			Ticker:   holding.Ticker,
			Volume:   sellVolume,
			NumTicks: holding.NumTicks,
			LastBid:  bid, LastAsk: ask,
			AvgBid: holding.AvgBid, AvgAsk: holding.AvgAsk,
			MaxBid: holding.MaxBid, MaxAsk: holding.MaxAsk,
			MinBid: holding.MinBid, MinAsk: holding.MinAsk,
			BuyPrice: holding.BuyPrice, SellPrice: ask,
		}
		// add new closed position to closedHoldings slice
		closedHoldings = append(closedHoldings, newClosedPosition)

		// Update total of transaction Amount
		transactionAmount += (ask.Amount * sellVolume)

		if holding.Volume == 0 {
			port.active[order.Ticker].pop(costMethod)
		}
		order.Volume -= sellVolume
	}
	oms.closeOutOrder(order)

	return transactionAmount, closedHoldings, nil
}

func (oms *OMS) closeOutOrder(closedOrder *Order) {
	orders := make([]*Order, 0)
	oms.Lock()
	for _, order := range oms.openOrders {
		if order != closedOrder {
			orders = append(orders, order)
		}
	}
	oms.openOrders = orders
	oms.Unlock()
}

// NewMarketOrder returns a new order that will execute at nearest price.
func NewMarketOrder(buy bool, ticker string, bid, ask, volume Amount, datetime time.Time) *Order {
	return &Order{
		Buy:      buy,
		Status:   open,
		Logic:    market,
		Ticker:   ticker,
		Bid:      bid,
		Ask:      ask,
		Volume:   volume,
		Datetime: datetime,
	}
}

// Order struct hold information referring to the
// details of an execution of a financial asset transaction.
type Order struct {
	// it's either a buy or sell
	Buy    bool
	Status OrderStatus
	Logic  TradeLogic
	Ticker string
	// NOTE: turn price + datetime into LastBid & LastAsk
	Bid, Ask, Volume Amount
	Datetime         time.Time
}

func (order *Order) toPosition() *Position {
	bid := &datedMetric{Amount: order.Bid, Date: order.Datetime}
	ask := &datedMetric{Amount: order.Ask, Date: order.Datetime}

	return &Position{
		Ticker:   order.Ticker,
		Volume:   order.Volume,
		NumTicks: 1,
		LastBid:  bid, LastAsk: ask,
		AvgBid: bid.Amount, AvgAsk: ask.Amount,
		MaxBid: bid, MaxAsk: ask,
		MinBid: bid, MinAsk: ask,
		BuyPrice: ask,
	}
}

// OrderStatus variables refer to a status of an order's execution.
type OrderStatus int

const (
	open OrderStatus = iota // 0
	closed
	cancelled
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
