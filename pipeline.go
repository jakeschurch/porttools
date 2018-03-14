package porttools

func newBacktestEngine(cashAmt Amount, toIgnore []string) *BacktestEngine {
	btEngine := &BacktestEngine{
		OMS: newOMS(cashAmt, toIgnore),
		// TODO: portLog
	}
	return btEngine
}

// BacktestEngine is the centralized struct that everything is occuring through within a simulation.
type BacktestEngine struct {
	Benchmark *Index
	Log       *PerformanceLog
	OMS       *OMS
}

// IDEA: send closed orders to PerformanceLog Closed slice instead of back to Portfolio's ClosedPosition Slice
// OPTIMIZE: instead of using sync.Mutexes, use channels/non-blocking functions

func newOMS(cashAmt Amount, toIgnore []string) *OMS {
	return &OMS{
		Portfolio:  NewPortfolio(cashAmt),
		OpenOrders: make([]*Order, 0),
		strategy:   newStrategy(toIgnore),
	}
}

// OMS acts as an `Order Management System` to test trading signals and fill orders.
type OMS struct {
	Portfolio  *Portfolio
	OpenOrders []*Order
	strategy   *Strategy
}

// newStrategy creates a new Strategy instance used in the backtesting process.
func newStrategy(toIgnore []string) *Strategy {
	toIgnoreMap := make(map[string]bool)
	for _, ticker := range toIgnore {
		toIgnoreMap[ticker] = true
	}
	strategy := &Strategy{ignoreTickers: toIgnoreMap}
	return strategy
}

// Strategy ... TODO
type Strategy struct {
	Algorithm
	ignoreTickers map[string]bool
}

// Algorithm is an interface that needs to be implemented in the pipeline by a user to fill orders based on the conditions that they specify.
type Algorithm interface {
	EntryLogic() bool
	ExitLogic() bool
	OrderLogic() bool
}

// PerformanceLog conducts performance analysis.
type PerformanceLog struct {
	Closed PositionSlice
	orders Queue
	// benchmark *Index
}

// TODO: ExecuteStrategy method

//
// import (
// 	"encoding/csv"
// 	"io"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"time"
// )

// option to throw out securities that aren't needed
// only see different trades -> simulate market that has traders doing different trades and what is the aggregate position look like
