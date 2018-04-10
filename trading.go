package porttools

import (
	"errors"

	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/order"
)

var (
	// ErrOrderNotValid indicates that an order is not valid, and should not be sent to the  pipeline.
	ErrOrderNotValid = errors.New("Order does not meet criteria as a valid order")
)

// Algorithm is an interface that needs to be implemented in the pipeline by a user to fill orders based on the conditions that they specify.
type Algorithm interface {
	EntryCheck(instrument.Quote) (*order.Order, error)
	ExitCheck(order.Order, instrument.Tick) (*order.Order, error)
}

// ------------------------------------------------------------------

// Strategy ...
type Strategy struct {
	Algorithm Algorithm
}

// NewStrategy creates a new Strategy instance used in the backtesting process.
func NewStrategy(a Algorithm) Strategy {
	return Strategy{
		Algorithm: a,
	}
}

// CheckEntryLogic ...TODO
func (s Strategy) CheckEntryLogic(q instrument.Quote) (entryOrder *order.Order, err error) {
	if entryOrder, err = s.Algorithm.EntryCheck(q); err != nil {
		return nil, ErrOrderNotValid
	}
	return entryOrder, nil
}

// CheckExitLogic ...TODO
func (s Strategy) CheckExitLogic(o order.Order, t instrument.Tick) (exitOrder *order.Order, err error) {
	if exitOrder, err = s.Algorithm.ExitCheck(o, t); err != nil {
		return nil, ErrOrderNotValid
	}
	return exitOrder, nil
}
