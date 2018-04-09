package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/jakeschurch/porttools"
	"github.com/jakeschurch/porttools/collection/portfolio"
	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/order"
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

var cpuprofile = flag.String("cpuprofile", "cpu.prof", "write cpu profile to file")
var memprofile = flag.String("memprofile", "mem.prof", "write memory profile to `file`")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	myAlgo := newAlgo()
	cfgFile := "/home/jake/go/src/github.com/jakeschurch/porttools/example/exampleConfig.json"
	sim, simErr := porttools.NewSimulation(*myAlgo, cfgFile)
	if simErr != nil {
		log.Fatal("Error in Simulation: ", simErr)
	}
	sim.Run()

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}
}
