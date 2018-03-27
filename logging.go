package porttools

import (
	"time"
)

func newPrfmLog(outfmt OutputFmt) *PrfmLog {
	return &PrfmLog{
		closedOrders:    make([]*Order, 0),
		closedPositions: make([]*Position, 0),
		results:         make([]*result, 0),
		// create channels.
		orderChan: make(chan *Order),
		posChan:   make(chan *Position),
		endMux:    make(chan bool, 1),
		outFmt:    outfmt,
	}
}

// PrfmLog conducts performance analysis.
type PrfmLog struct {
	closedOrders    []*Order
	closedPositions []*Position
	results         []*result
	orderChan       chan *Order
	posChan         chan *Position
	endMux          chan bool
	outFmt          OutputFmt
}

func (prfmLog *PrfmLog) mux() {
	done := make(chan struct{})

	for prfmLog.orderChan != nil || prfmLog.posChan != nil {
		select {
		case order, ok := <-prfmLog.orderChan:
			if !ok {
				prfmLog.orderChan = nil
				continue
			}
			prfmLog.closedOrders = append(prfmLog.closedOrders, order)

		case pos, ok := <-prfmLog.posChan:
			if !ok {
				prfmLog.posChan = nil
				continue
			}
			prfmLog.closedPositions = append(prfmLog.closedPositions, pos)

		case <-prfmLog.endMux:
			done <- struct{}{}

		default:
			time.Sleep(1 * time.Nanosecond)
		}
	}
	<-done
}

func (prfmLog *PrfmLog) quit() {

	close(prfmLog.orderChan)
	close(prfmLog.posChan)
	prfmLog.endMux <- true
}

func (oms *OMS) getResults() {
	// sort orders by ticker for easier access
	// selectionSort(oms.prfmLog.closedPositions)

	// create slice of all position keys
	positionKeys := make([]string, 0)

	for _, position := range oms.prfmLog.closedPositions {
		if !findKey(positionKeys, position.Ticker) {
			positionKeys = append(positionKeys, position.Ticker)
		}
	}
	for _, key := range positionKeys {
		filtered := filter(oms.prfmLog.closedPositions, key)
		oms.prfmLog.results = append(oms.prfmLog.results, oms.createResult(filtered))
	}
	oms.outputResults()
	oms.end <- struct{}{}
}

func (oms *OMS) createResult(positions []*Position) *result {
	// create result struct
	result := &result{Ticker: positions[0].Ticker}

	// loop through and update metrics accordingly
	for _, pos := range positions {
		result.update(pos)
	}
	result.averageize()

	security, ok := oms.benchmark.Instruments[result.Ticker]
	if !ok {
		result.Alpha = Amount(0)
		return result
	}
	// NOTE: this is NOT on an aggregate basis at the moment.
	alpha := Amount(result.PctReturn) - (security.LastAsk.Amount-security.BuyPrice.Amount)/security.BuyPrice.Amount
	result.Alpha = alpha
	return result
}

// - max-drawdown
// - % profitable
// - total num trades
// - winning/losing trades
// - trading period length

// only see different trades -> simulate market that has traders doing different trades and what their aggregate position look like
