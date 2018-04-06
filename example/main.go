package main

import (
	"log"

	pt "github.com/jakeschurch/porttools"
)

func newAlgo() *algo {
	return &algo{}
}

type algo struct{}

func (algo algo) EntryLogic(tick pt.Tick) (*pt.Order, bool) {
	// if tick.Ticker == "AAPL" {

	// 	pt.DivideAmt(tick.Ask-tick.Bid, tick.Ask) <= 2 &&
	// 	tick.AskSize <= pt.FloatAmount(50.00) {

	return pt.NewMarketOrder(
		true, tick.Ticker, tick.Bid, tick.Ask,
		pt.Amount(50.00), tick.Timestamp), true
	// }
	// return nil, false
}

func (algo algo) ExitLogic(tick pt.Tick, openOrder *pt.Order) (*pt.Order, bool) {
	// pctMoved := pt.DivideAmt(tick.Ask-openOrder.Bid, openOrder.Bid)
	// if pctMoved >= 1 || pctMoved <= -1 {
	// 	return pt.NewMarketOrder(
	// 		false, tick.Ticker, tick.Bid, tick.Ask, openOrder.Volume, tick.Timestamp), true
	// }
	// return nil, false
	if tick.Ticker == openOrder.Ticker {
		return pt.NewMarketOrder(
			false, tick.Ticker, tick.Bid, tick.Ask, openOrder.Volume, tick.Timestamp), true
	}
	return nil, false
}

func (algo algo) ValidOrder(port *pt.Portfolio, order *pt.Order) bool {
	// port.RLock()
	// defer port.RUnlock()

	// if order.Buy == true {
	// 	cashLeft := port.Cash - (order.Ask * order.Volume)
	// 	if cashLeft/pt.FloatAmount(100) >= pt.FloatAmount(50000.00) {
	// 		return true
	// 	}
	// 	return false
	// }
	return true
}

func main() {

	myAlgo := newAlgo()
	cfgFile := "/home/jake/go/src/github.com/jakeschurch/porttools/example/exampleConfig.json"
	sim, simErr := pt.NewSimulation(*myAlgo, cfgFile)
	if simErr != nil {
		log.Fatal("Error in Simulation: ", simErr)
	}
	sim.Run()
}
