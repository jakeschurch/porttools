package porttools

func newPerfLog() *PerfLog {
	return &PerfLog{
		closedOrders:    make([]*Order, 0),
		closedPositions: make([]*Position, 0),
		results:         make([]*result, 0),
		// create channels.
		orderChan: make(chan *Order),
		posChan:   make(chan *Position),
		startQuit: make(chan bool, 1),
	}
}

// PerfLog conducts performance analysis.
type PerfLog struct {
	closedOrders    []*Order
	closedPositions []*Position
	results         []*result
	orderChan       chan *Order
	posChan         chan *Position
	startQuit       chan bool // TODO: comon, you can do better than that XD
}

func (log *PerfLog) run() {
	done := make(chan struct{})

	go func() {
		for log.orderChan != nil || log.posChan != nil {
			select {
			case order, ok := <-log.orderChan:
				if !ok {
					log.orderChan = nil
					continue
				}
				log.closedOrders = append(log.closedOrders, order)

			case pos, ok := <-log.posChan:
				if !ok {
					log.posChan = nil
					continue
				}
				log.closedPositions = append(log.closedPositions, pos)

			case <-log.startQuit:
				done <- struct{}{}
			}
		}
	}()
	<-done
}

func (log *PerfLog) quit() {
	close(log.orderChan)
	close(log.posChan)
	log.startQuit <- true
}

func (oms *OMS) getResults() {
	// sort orders by ticker for easier access
	selectionSort(oms.log.closedPositions)

	// create slice of all position keys
	positionKeys := make([]string, 0)

	for _, position := range oms.log.closedPositions {
		if !findKey(positionKeys, position.Ticker) {
			positionKeys = append(positionKeys, position.Ticker)
		}
	}
	for _, key := range positionKeys {
		filtered := filter(oms.log.closedPositions, key)
		oms.log.results = append(oms.log.results, oms.createResult(filtered))
	}
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
		result.Alpha = result.PctReturn
		return result
	}
	// NOTE: this is NOT on an aggregate basis at the moment.
	alpha := Amount(result.PctReturn) - (security.LastAsk.Amount-security.BuyPrice.Amount)/security.BuyPrice.Amount
	result.Alpha = float64(alpha / 100)
	return result

}

type result struct {
	Ticker    string `json:"ticker"`
	Filled    uint   `json:"filled"`
	AvgVolume Amount `json:"avgVolume"`

	BuyValue Amount `json:"buyValue"`
	EndValue Amount `json:"endValue"`

	AvgBid Amount      `json:"avgBid"`
	MaxBid datedMetric `json:"maxBid"`
	MinBid datedMetric `json:"minBid"`

	AvgAsk Amount      `json:"avgAsk"`
	MaxAsk datedMetric `json:"maxAsk"`
	MinAsk datedMetric `json:"minAsk"`
	// REVIEW: avgReturn
	PctReturn float64 `json:"pctReturn"`
	Alpha     float64 `json:"alpha"`
}

func (result *result) update(pos *Position) {
	result.AvgBid += pos.AvgBid
	result.AvgAsk += pos.AvgAsk

	result.BuyValue += pos.BuyPrice.Amount * pos.Volume
	result.EndValue += pos.SellPrice.Amount * pos.Volume

	result.Filled++

	result.MaxBid = newMax(result.MaxBid, pos.MaxBid.Amount, pos.MaxBid.Date)
	result.MinBid = newMin(result.MinBid, pos.MinBid.Amount, pos.MinBid.Date)

	result.MaxAsk = newMax(result.MaxAsk, pos.MaxAsk.Amount, pos.MaxAsk.Date)
	result.MinAsk = newMin(result.MinAsk, pos.MinAsk.Amount, pos.MinAsk.Date)
	return
}

func (result *result) averageize() { // REVIEW:  oh dear god
	amtFilled := Amount(result.Filled)
	result.AvgBid /= amtFilled
	result.AvgAsk /= amtFilled

	result.PctReturn = float64((result.EndValue - result.BuyValue) / result.BuyValue)
	return
}

// - max-drawdown
// - % profitable
// - total num trades
// - winning/losing trades
// - trading period length

// only see different trades -> simulate market that has traders doing different trades and what their aggregate position look like
