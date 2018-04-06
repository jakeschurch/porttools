package porttools

import (
	"sync"

	"github.com/jakeschurch/porttools/instrument/holding"
	"github.com/jakeschurch/porttools/trading/order"
)

// PrfmLog allows for performance analysis.
type PrfmLog struct {
	sync.Mutex
	ClosedOrders   []*order.Order
	ClosedHoldings []holding.Holding
	orderChan      chan *order.Order
	holdingChan    chan *holding.Holding
	errChan        chan error
	endMux         chan bool
}

func newPrfmLog() *PrfmLog {
	prfmLog := PrfmLog{
		ClosedOrders:   make([]*order.Order, 0),
		ClosedHoldings: make([]holding.Holding, 0),
	}
	return &prfmLog
}

// AddHolding adds a closed holding to the performance log's closed holdings slice.
func (prfmLog *PrfmLog) AddHolding(holding holding.Holding) error {
	prfmLog.Lock()
	prfmLog.ClosedHoldings = append(prfmLog.ClosedHoldings, holding)
	prfmLog.Unlock()
	return nil
}

// AddOrder adds a closed order to the performance log's closed order slice.
func (prfmLog *PrfmLog) AddOrder(order *order.Order) error {
	prfmLog.ClosedOrders = append(prfmLog.ClosedOrders, order)
	return nil
}
