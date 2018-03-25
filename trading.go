package porttools

import (
	"errors"
	"log"
)

// Algorithm is an interface that needs to be implemented in the pipeline by a user to fill orders based on the conditions that they specify.
type Algorithm interface {
	// REVIEW: may want to move this to pipeline || simulation.
	EntryLogic(*Tick) (*Order, bool)
	ExitLogic(*Tick, *Order) (*Order, bool)
	ValidOrder(*OMS, *Order) bool
}

// CostMethod regards the type of accounting management rule
// is implemented for selling securities.
type CostMethod int

const (
	lifo CostMethod = iota - 1
	fifo
)

// Transact conducts agreement between Position and Order within a portfolio.
func (oms *OMS) Transact(order *Order) error {
	log.Println("Transacting order...")
	switch order.Buy {
	case true: // in lieu of a buy function
		ok := func() bool { // check to see if order can be fulfilled.
			oms.Port.RLock()
			ok := (order.Volume * order.Price) <= oms.Port.Cash
			oms.Port.RUnlock()
			return ok
		}()
		if !ok { // if not ok, cancel order and return error.
			order.Status = cancelled
			go func() { oms.prfmLog.orderChan <- order }()
			return errors.New("Not enough cash to fulfil order")
		}
		if _, exists := oms.Port.Active[order.Ticker]; !exists {
			oms.Port.Lock()
			oms.Port.Active[order.Ticker] = newPositionSlice()
			oms.Port.Unlock()
		}
		order.Status = open
		// Create new Position and add it to according position slice.
		posBought := order.toPosition(order.Volume)
		oms.Port.Active[order.Ticker].Push(posBought)
		oms.Port.Active[order.Ticker].totalVolume += posBought.Volume // Update position slice volume.
		go func() { oms.openChan <- order }()

	case false: // sell
		// Check to see if order can be fulfilled
		// if not, cancel order and return error
		if oms.Port.Active[order.Ticker].totalVolume < order.Volume {
			order.Status = cancelled
			go func() { oms.prfmLog.orderChan <- order }()

			return errors.New("Not enough volume to satisfy order")
		}
		order.Status = closed
		go func() {
			oms.prfmLog.orderChan <- order
			oms.sell(*order)
		}()
	}
	return nil
}

// sell is a function that removes a Position's volume, as well as create
// a new closed position. Updates a port's cash balance.
func (oms *OMS) sell(order Order) (err error) {
	// Update Cash Amount.
	oms.Port.Lock()
	oms.Port.Cash += order.Volume * order.Price
	oms.Port.Unlock()

	for order.Volume > 0 {
		var posToSell *Position
		if posToSell = oms.Port.Active[order.Ticker].Peek(oms.strat.costMethod); posToSell == nil {
			err = errors.New("No position to sell")
			return
		}

		var sellVolume Amount
		if posToSell.Volume >= order.Volume {
			sellVolume = order.Volume
		} else {
			sellVolume = posToSell.Volume
		}
		posToSell.Volume -= sellVolume

		closedPos := *posToSell
		closedPos.Volume = sellVolume
		closedPos.SellPrice = datedMetric{order.Price, order.Datetime}
		go func() {
			oms.prfmLog.posChan <- &closedPos
		}()

		if posToSell.Volume == 0 {
			_, popErr := oms.Port.Active[order.Ticker].Pop(oms.strat.costMethod)
			if popErr != nil {
				err = popErr
				return
			}
		}
		if oms.Port.Active[order.Ticker].len == 0 {
			oms.Port.Lock()
			delete(oms.Port.Active, order.Ticker)
			oms.Port.Unlock()
		}
	}
	return
}
