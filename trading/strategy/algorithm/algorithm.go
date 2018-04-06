package algorithm

import (
	"github.com/jakeschurch/porttools/collection/portfolio"
	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/trading/order"
)

// Algorithm is an interface that needs to be implemented in the pipeline by a user to fill orders based on the conditions that they specify.
type Algorithm interface {
	// REVIEW: may want to move this to pipeline || simulation.
	EntryLogic(instrument.Tick) (*order.Order, bool)
	ExitLogic(instrument.Tick, *order.Order) (*order.Order, bool)
	ValidOrder(*portfolio.Portfolio, *order.Order) bool
}
