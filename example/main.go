package main

import (
	"log"

	"github.com/jakeschurch/porttools"
	"github.com/jakeschurch/porttools/collection/portfolio"
	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/trading/order"
	"github.com/jakeschurch/porttools/utils"
)

func newAlgo() *algo {
	return &algo{}
}

type algo struct{}

func (algo algo) EntryLogic(tick instrument.Tick) (*order.Order, bool) {
	// if tick.Ticker == "AAPL" {

	// 	pt.DivideAmt(tick.Ask-tick.Bid, tick.Ask) <= 2 &&
	// 	tick.AskSize <= pt.FloatAmount(50.00) {

	return order.New(
		true, tick.Ticker, tick.Bid, tick.Ask,
		utils.Amount(50.00), tick.Timestamp), true
	// }
	// return nil, false
}

func (algo algo) ExitLogic(tick instrument.Tick, openOrder *order.Order) (*order.Order, bool) {
	// pctMoved := pt.DivideAmt(tick.Ask-openOrder.Bid, openOrder.Bid)
	// if pctMoved >= 1 || pctMoved <= -1 {
	// 	return pt.NewMarketOrder(
	// 		false, tick.Ticker, tick.Bid, tick.Ask, openOrder.Volume, tick.Timestamp), true
	// }
	// return nil, false
	if tick.Ticker == openOrder.Ticker {
		return order.New(
			false, tick.Ticker, tick.Bid, tick.Ask, openOrder.Volume, tick.Timestamp), true
	}
	return nil, false
}

func (algo algo) ValidOrder(port *portfolio.Portfolio, order *order.Order) bool {
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
	sim, simErr := porttools.NewSimulation(*myAlgo, cfgFile)
	if simErr != nil {
		log.Fatal("Error in Simulation: ", simErr)
	}
	sim.Run()
}
