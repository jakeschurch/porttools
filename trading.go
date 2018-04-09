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

type EntryCheck func(instrument.Quote, ...QueryFunc) (*order.Order, error)
type ExitCheck func(order.Order, ...QueryFunc) (*order.Order, error)

// Algorithm is an interface that needs to be implemented in the pipeline by a user to fill orders based on the conditions that they specify.
type Algorithm interface {
	EntryCheck() EntryCheck
	ExitCheck() ExitCheck
}

// ------------------------------------------------------------------

// Strategy ...
type Strategy struct {
	Algorithm             Algorithm
	entryQuery, exitQuery QueryFunc
}

// NewStrategy creates a new Strategy instance used in the backtesting process.
func NewStrategy(a Algorithm, entryQuery, exitQuery QueryFunc) Strategy {
	return Strategy{
		Algorithm:  a,
		entryQuery: entryQuery,
		exitQuery:  exitQuery,
	}
}

// CheckEntryLogic ...TODO
func (s Strategy) CheckEntryLogic(q *instrument.Quote) (*order.Order, error) {
	if order, err := s.EntryCheck(q, entryQuery); err != nil {
		return nil, ErrOrderNotValid
	}
	return order, nil
}

// CheckExitLogic ...TODO
func (s Strategy) CheckExitLogic(o order.Order) (*order.Order, error) {
	if order, err := s.ExitCheck(o, exitQuery); err != nil {
		return nil, ErrOrderNotValid
	}
	return order, nil
}
