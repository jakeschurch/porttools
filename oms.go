package porttools

import (
	"errors"
	"sync"

	"github.com/jakeschurch/porttools/collection"
	ins "github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/utils"
)

var (
	// ErrNegativeVolume indicates that the processed order
	// has too high of a volume to act upon.
	ErrNegativeVolume = errors.New("not enough volume to fill order")
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

// Insert checks to see if we can insert a new buy order into the OMS.
// If it can, order will be inserted into oms, updates cash,
// and stores new holding in Port.
func (oms *OMS) Insert(o *ins.Order) error {
	var dxCash utils.Amount

	switch o.Buy {
	case true:
		dxCash = -o.Ask * o.Volume

	case false:
		dxCash = o.Bid * o.Volume
	}
	if err := oms.open.Insert(o); err != nil {
		return err
	}
	oms.updateCash(dxCash)
	i := ins.NewHolding(*o.Quote)
	return Port.Insert(i, *o.Quote)
}

func (oms *OMS) Query(t *ins.Tick) error {
	var entryOrder *ins.Order

	switch entryOrder, _ = strategy.CheckEntryLogic(*t.Quote); entryOrder != nil {
	case true:
		oms.Insert(entryOrder)
	case false: // do nothing if entry logic is not met.
	}
	return oms.queryOpenOrders(*t)
}

func (oms *OMS) queryOpenOrders(t ins.Tick) error {
	var orderList *collection.LinkedList
	var openOrderNode *collection.LinkedNode
	var exitOrder *ins.Order
	var err error

	if orderList, err = oms.open.Get(t.Ticker); err != nil {
		return err
	}

	for openOrderNode = orderList.PeekFront(); openOrderNode != nil; openOrderNode = openOrderNode.Next() {

		// TEMP: for now, do nothing with exitOrder
		exitOrder, err = strategy.CheckExitLogic(openOrderNode.Data.(ins.Order), t)

		switch err != nil {
		case false:
			if err = oms.open.RemoveNode(openOrderNode); err != nil {
				return err
			}
			oms.updateCash(exitOrder.Volume * exitOrder.Bid)

		case true: // do nothing if invalid exit logic
		}
	}
	return nil
}

func (oms *OMS) updateCash(dxCash utils.Amount) {
	oms.mu.Lock()
	oms.cash += dxCash
	oms.mu.Unlock()

	return
}

func (oms *OMS) executeSell(o ins.Order) error {
	var closed = make([]*ins.Security, 0)
	var list *collection.LinkedList
	var closedSecurity *ins.Security
	var toSell *collection.LinkedNode
	var sellVolume utils.Amount
	var err error

	var orderVolume = o.Volume
	var ticker = o.Ticker

	if list, err = Port.GetList(ticker); err != nil {
		return err
	}

	if list.TotalVolume < o.Volume {
		return ErrNegativeVolume
	}

	// loop over slice until order has been completely crumpled
	for orderVolume > 0 {
		toSell = Port.Peek(ticker, costMethod)
		holdingVolume := toSell.Data.(ins.Quote).Volume

		switch holdingVolume >= orderVolume {
		case true:
			sellVolume = orderVolume
		case false:
			sellVolume = holdingVolume
		}

		if closedSecurity, err = list.PeekToSecurity(costMethod, o); err != nil {
			return nil
		}

		closed = append(closed, closedSecurity)

		if toSell.Data.(ins.Quote).Volume == 0 {
			list.Pop(costMethod)
		}
		orderVolume -= sellVolume
	}
	return positionLog.Insert(closed...)
}

func (oms *OMS) Cash() utils.Amount {
	oms.mu.RLock()
	cash := oms.cash
	oms.mu.RUnlock()
	return cash
}
