package porttools

import (
	"sync"

	"github.com/jakeschurch/porttools/instrument"

	"github.com/jakeschurch/porttools/collection"

	"github.com/jakeschurch/porttools/trading/order"
)

// PrfmLog allows for performance analysis.
type PrfmLog struct {
	sync.Mutex
	ClosedPositions collection.HoldingList
}

func newPrfmLog() *PrfmLog {
	p := PrfmLog{
		ClosedPositions: collection.NewHoldingList(),
	}
	return &p
}

// AddHolding adds a closed holding to the performance log's closed holdings slice.
func (p *PrfmLog) Insert(s *instrument.Security) error {
	return p.ClosedPositions.Insert(s)
}

// AddOrder adds a closed order to the performance log's closed order slice.
func (p *PrfmLog) AddOrder(order *order.Order) error {
	prfmLog.ClosedOrders = append(prfmLog.ClosedOrders, order)
	return nil
}
