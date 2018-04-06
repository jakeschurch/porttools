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
func newStrategy(algo Algorithm, toIgnore []string) Strategy {
	strategy := Strategy{
		algorithm: algo,
		ignore:    toIgnore,
	}
	return strategy
}

// Strategy ... TODO
type Strategy struct {
	algorithm Algorithm
	ignore    []string // TODO: REVIEW later...
}

func (strategy Strategy) checkEntryLogic(port *Portfolio, tick Tick) (*Order, error) {
	if order, signal := strategy.algorithm.EntryLogic(tick); signal {
		if strategy.algorithm.ValidOrder(port, order) {
			return order, nil
		}
	}
	return nil, ErrOrderNotValid
}

func (strategy Strategy) checkExitLogic(port *Portfolio, openOrder *Order, tick Tick) (*Order, error) {
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
