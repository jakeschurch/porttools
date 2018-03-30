package porttools

import (
	"errors"
)

var (
	// ErrOrderNotValid indicates that an order is not valid, and should not be sent further in the pipeline.
	ErrOrderNotValid = errors.New("Order does not meet criteria as a valid order")
)

// Algorithm is an interface that needs to be implemented in the pipeline by a user to fill orders based on the conditions that they specify.
type Algorithm interface {
	// REVIEW: may want to move this to pipeline || simulation.
	EntryLogic(Tick) (*Order, bool)
	ExitLogic(Tick, *Order) (*Order, bool)
	ValidOrder(*Portfolio, *Order) bool
}

// newStrategy creates a new Strategy instance used in the backtesting process.
func newStrategy(algo Algorithm, toIgnore []string) strategy {
	strategy := strategy{
		algorithm: algo,
		ignore:    toIgnore,
	}
	return strategy
}

// strategy ... TODO
type strategy struct {
	algorithm Algorithm
	ignore    []string // TODO: REVIEW later...
}

func (strategy strategy) checkEntryLogic(port *Portfolio, tick Tick) (*Order, error) {
	if order, signal := strategy.algorithm.EntryLogic(tick); signal {
		if strategy.algorithm.ValidOrder(port, order) {
			return order, nil
		}
	}
	return nil, ErrOrderNotValid
}

func (strategy strategy) checkExitLogic(port *Portfolio, openOrder *Order, tick Tick) (*Order, error) {
	if order, signal := strategy.algorithm.ExitLogic(tick, openOrder); signal {
		if strategy.algorithm.ValidOrder(port, order) {
			return order, nil
		}
	}
	return nil, ErrOrderNotValid
}

// CostMethod regards the type of accounting management rule
// is implemented for selling securities.
type CostMethod int

const (
	lifo CostMethod = iota - 1
	fifo
)

// // Transact conducts agreement between Position and Order within a portfolio.
// func (oms *OMS) Transact(orderChan chan<- *Order, order *Order) error {

// 	switch order.Buy {
// 	case true: // in lieu of a buy function
// 		oms.Port.Lock()
// 		defer oms.Port.Unlock()
// 		if _, exists := oms.Port.active[order.Ticker]; !exists {
// 			oms.Port.active[order.Ticker] = NewHoldingSlice()
// 		}
// 		order.Status = open
// 		// Create new Position and add it to according position slice.
// 		posBought := order.toPosition(order.Volume)
// 		oms.Port.active[order.Ticker].Push(posBought)

// 		oms.openOrders = append(oms.openOrders, order)

// 	case false: // sell
// 		// Check to see if order can be fulfilled
// 		// if not, change order volume to max amount it can sell
// 		if oms.Port.active[order.Ticker].totalVolume < order.Volume {
// 			order.Status = cancelled

// 			oms.prfmLog.closedOrders = append(oms.prfmLog.closedOrders, order)
// 			log.Println("Not enough volume to satisfy order")
// 			return errors.New("Not enough volume to satisfy order")
// 		}
// 		order.Status = closed
// 		oms.prfmLog.closedOrders = append(oms.prfmLog.closedOrders, order)
// 		oms.sell(*order)
// 	}
// 	return nil
// }

// // sell is a function that removes a Position's volume, as well as create
// // a new closed position. Updates a port's cash balance.
// // TODO: remove this after Sim function for deleting holdingSlice done.
// func (oms *OMS) sell(order Order) (err error) {
// 	// Update Cash Amount.
// 	oms.Port.Lock()
// 	oms.Port.cash += order.Volume * order.Ask
// 	oms.Port.Unlock()

// 	for order.Volume > 0 {
// 		var posToSell *Position
// 		if posToSell = oms.Port.active[order.Ticker].Peek(oms.strat.costMethod); posToSell == nil {
// 			err = errors.New("No position to sell")
// 			return
// 		}

// 		var sellVolume Amount
// 		if posToSell.Volume >= order.Volume {
// 			sellVolume = order.Volume
// 		} else {
// 			sellVolume = posToSell.Volume
// 		}
// 		posToSell.Volume -= sellVolume

// 		closedPos := *posToSell
// 		closedPos.Volume = sellVolume
// 		closedPos.SellPrice = &datedMetric{order.Bid, order.Datetime}
// 		oms.prfmLog.posChan <- &closedPos

// 		if posToSell.Volume == 0 {
// 			_, popErr := oms.Port.active[order.Ticker].Pop(oms.strat.costMethod)
// 			if popErr != nil {
// 				err = popErr
// 				return
// 			}
// 		}
// 		if oms.Port.active[order.Ticker].len == 0 {
// 			oms.Port.Lock()
// 			delete(oms.Port.active, order.Ticker)
// 			oms.Port.Unlock()
// 		}
// 	}
// 	return
// }
