package porttools

// TODO: Finish method signature
type algorithm interface {
	EntryLogic() bool
	ExitLogic() bool
}

// TODO: Algorithm ...
type Algorithm struct{}

// TODO: Strategy ...
type Strategy struct {
	algos []Algorithm
}

/* TODO:
	- Restricted tickers
 	- MaxOrderSize
	- MaxPositionSize
	- LongOnly/ShortOnly
	- AssetDateBounds
*/

// TODO: researchSys ...
type researchSys struct {
	Strategy Strategy
	tickChan chan Tick
}

// TODO: ExecuteStrategy method
