package main

import (
	"log"

	"github.com/jakeschurch/porttools"
	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/utils"
)

func newAlgo() *algo {
	return &algo{}
}

type algo struct{}

func (algo algo) EntryCheck(q instrument.Quote) (*instrument.Order, error) {
	var cash = porttools.Oms.Cash()

	if cash-q.Ask*q.Volume(0) < 0 {
		return nil, porttools.ErrOrderNotValid
	}

	newOrder := instrument.NewOrder(true, q)

	newOrder.Volume(utils.Amount(50.00))
	return newOrder, nil
}

func (algo algo) ExitCheck(openOrder instrument.Order, t instrument.Tick) (*instrument.Order, error) {
	if t.Ticker() == openOrder.Ticker() {
		return instrument.NewOrder(false, t.Quote), nil
	}
	return nil, porttools.ErrOrderNotValid
}

// var cpuprofile = flag.String("cpuprofile", "cpu.prof", "write cpu profile to file")
// var memprofile = flag.String("memprofile", "mem.prof", "write memory profile to `file`")

func main() {
	// log.Println("starting")
	// flag.Parse()
	// if *cpuprofile != "" {
	// 	f, err := os.Create(*cpuprofile)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	pprof.StartCPUProfile(f)
	// 	defer pprof.StopCPUProfile()
	// }

	cfgFile := "/home/jake/go/src/github.com/jakeschurch/porttools/example/exampleConfig.json"
	log.Println("running strategy")
	porttools.NewStrategy(newAlgo())
	sim, simErr := porttools.NewSimulation(cfgFile)

	if simErr != nil {
		log.Fatal("Error in Simulation: ", simErr)
	}
	log.Println("running sim")
	sim.Run()

	// if *memprofile != "" {
	// 	f, err := os.Create(*memprofile)
	// 	if err != nil {
	// 		log.Fatal("could not create memory profile: ", err)
	// 	}
	// 	runtime.GC() // get up-to-date statistics
	// 	if err := pprof.WriteHeapProfile(f); err != nil {
	// 		log.Fatal("could not write memory profile: ", err)
	// 	}
	// 	f.Close()
	// }
}
