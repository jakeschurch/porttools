package main

import (
	pt "github.com/jakeschurch/porttools"
	"log"
)

func newAlgo() *algo {
	return &algo{}
}

type algo struct{}

func (algo *algo) EntryLogic(tick *pt.Tick) (*pt.Order, bool) {
	if tick.Ticker == "AAPL" {
		return pt.NewMarketOrder(
			true, tick.Ticker, tick.Bid,
			pt.FloatAmount(100.00), tick.Timestamp), true
	}
	return nil, false
}

func (algo *algo) ExitLogic(tick *pt.Tick, openOrder *pt.Order) (*pt.Order, bool) {
	// 100 being 1.00% if it was of the porttools Amount type.
	if (tick.Bid-openOrder.Price)/openOrder.Price >= 150 {
		return pt.NewMarketOrder(
			false, tick.Ticker, tick.Bid, openOrder.Volume, tick.Timestamp), true
	}
	return nil, false
}

func (algo *algo) ValidOrder(oms *pt.OMS, order *pt.Order) bool {
	if cashLeft := oms.Port.Cash - (order.Price * order.Volume); cashLeft > pt.FloatAmount(50000.00) {
		return true
	}
	return false
}

func main() {

	myAlgo := newAlgo()
	cfgFile := "/home/jake/code/go/workspace/src/github.com/jakeschurch/porttools/example/exampleConfig.json"
	sim, simErr := pt.NewSimulation(cfgFile)
	if simErr != nil {
		log.Fatal("Something went wrong", simErr)
	}
	pt.LoadAlgorithm(sim, myAlgo)
}
