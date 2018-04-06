package porttools

import (
	"log"
	"sync"
)

// PrfmLog allows for performance analysis.
type PrfmLog struct {
	sync.Mutex
	closedOrders    []*Order
	closedPositions []Position
	orderChan       chan *Order
	posChan         chan *Position
	errChan         chan error
	endMux          chan bool
}

func newPrfmLog() *PrfmLog {
	prfmLog := PrfmLog{
		closedOrders:    make([]*Order, 0),
		closedPositions: make([]Position, 0),
	}
	return &prfmLog
}

// AddPosition adds a closed position to the performance log's closed position slice.
func (prfmLog *PrfmLog) addPosition(pos Position) error {
	prfmLog.Lock()
	prfmLog.closedPositions = append(prfmLog.closedPositions, pos)
	prfmLog.Unlock()
	return nil
}

// AddOrder adds a closed order to the performance log's closed order slice.
func (prfmLog *PrfmLog) addOrder(order *Order) error {
	prfmLog.closedOrders = append(prfmLog.closedOrders, order)
	return nil
}

func getResults(closedPositions []Position, benchmarkMap map[string]*Security, outputFormat OutputFmt) {
	var results []result

	// create slice of all position keys
	positionKeys := make([]string, 0)

	for _, position := range closedPositions {

		if !findKey(positionKeys, position.Ticker) {
			positionKeys = append(positionKeys, position.Ticker)
		}
	}
	for _, key := range positionKeys {
		filtered := filter(closedPositions, key)
		results = append(results, createResult(filtered, benchmarkMap[key]))
	}

	log.Println("Outputting results: ")
	switch outputFormat {
	case CSV:
		resultsToCSV(results)
	}
}

func newResult(pos Position) result {
	result := result{
		Ticker:    pos.Ticker,
		Filled:    1,
		AvgVolume: pos.Volume,

		BuyValue: pos.BuyPrice.Amount * pos.Volume,
		EndValue: pos.SellPrice.Amount * pos.Volume,

		AvgBid: pos.AvgBid,
		MaxBid: pos.MaxBid,
		MinBid: pos.MinBid,

		AvgAsk: pos.AvgAsk,
		MaxAsk: pos.MaxAsk,
		MinAsk: pos.MinAsk,
	}
	return result
}
func createResult(positions []Position, security *Security) result {
	var posResult result

	// loop through and update metrics accordingly
	for index, pos := range positions {
		if index == 0 {
			posResult = newResult(positions[index])
		}
		posResult.update(pos)
	}
	posResult.averageize()

	if security == nil {
		posResult.Alpha = Amount(0)
		return posResult
	}
	// NOTE: this is NOT on an aggregate basis at the moment.
	benchmarkReturn := DivideAmt((security.LastAsk.Amount - security.BuyPrice.Amount), security.BuyPrice.Amount)
	posResult.Alpha = posResult.PctReturn - benchmarkReturn
	return posResult
}

// - max-drawdown
// - % profitable
// - total num trades
// - winning/losing trades
// - trading period length

// only see different trades -> simulate market that has traders doing different trades and what their aggregate position look like
