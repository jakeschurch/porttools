package porttools

// newStrategy creates a new Strategy instance used in the backtesting process.
func newStrategy(toIgnore []string, costMethod CostMethod) *strategy {
	toIgnoreMap := make(map[string]bool)
	for _, ticker := range toIgnore {
		toIgnoreMap[ticker] = true
	}
	strategy := &strategy{
		ignore:     toIgnoreMap,
		costMethod: costMethod,
	}
	return strategy
}

// strategy ... TODO
type strategy struct {
	algo       Algorithm
	costMethod CostMethod
	ignore     map[string]bool // TEMP: change this to function that can be embedded in early part of pipeline.
}
