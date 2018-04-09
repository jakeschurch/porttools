package strategy

import (
	"github.com/jakeschurch/porttools/collection/portfolio"
	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/trading"
	"github.com/jakeschurch/porttools/trading/order"
)

// Strategy ... TODO
type Strategy struct {
	Algorithm porttools.Algorithm
	ignore    []string // TODO: REVIEW later...
}

// New creates a new Strategy instance used in the backtesting process.
func New(algo algorithm.Algorithm, toIgnore []string) Strategy {
	strategy := Strategy{
		Algorithm: algo,
		ignore:    toIgnore,
	}
	return strategy
}

// CheckEntryLogic ...TODO
func (strategy Strategy) CheckEntryLogic(port *portfolio.Portfolio, tick instrument.Tick) (*order.Order, error) {
	if order, signal := strategy.Algorithm.EntryLogic(tick); signal {
		if strategy.Algorithm.ValidOrder(port, order) {
			return order, nil
		}
	}
	return nil, trading.ErrOrderNotValid
}

// CheckExitLogic ...TODO
func (strategy Strategy) CheckExitLogic(port *portfolio.Portfolio, openOrder *order.Order, tick instrument.Tick) (*order.Order, error) {
	if order, signal := strategy.Algorithm.ExitLogic(tick, openOrder); signal {
		if strategy.Algorithm.ValidOrder(port, order) {
			return order, nil
		}
	}
	return nil, trading.ErrOrderNotValid
}
