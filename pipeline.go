package porttools

func newBacktestEngine(cashAmt Amount, costMethod CostMethod, toIgnore []string) *BacktestEngine {
	btEngine := &BacktestEngine{
		OMS: newOMS(cashAmt, costMethod, toIgnore),
		// TODO: portLog
	}
	return btEngine
}

// BacktestEngine is the centralized struct that everything is occuring through within a simulation.
// REVIEW: is this struct needed? Should I promote OMS instead in Simulation struct?
type BacktestEngine struct {
	Log *PerfLog
	OMS *OMS
}

// newStrategy creates a new Strategy instance used in the backtesting process.
func newStrategy(toIgnore []string, costMethod CostMethod) *strategy {
	toIgnoreMap := make(map[string]bool)
	for _, ticker := range toIgnore {
		toIgnoreMap[ticker] = true
	}
	strat := &strategy{
		ignore:     toIgnoreMap,
		costMethod: costMethod,
	}
	return strat
}

// strategy ... TODO
type strategy struct {
	algo       Algorithm
	costMethod CostMethod
	ignore     map[string]bool // TEMP: change this to function that can be embedded in early part of pipeline.
}
