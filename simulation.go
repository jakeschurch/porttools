package porttools

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// LoadAlgorithm ensures that an Algorithm interface is implemented in the Simulation pipeline to be used by other functions.
func LoadAlgorithm(sim *Simulation, algo Algorithm) bool {
	sim.oms.strat.algo = algo
	return true
}

// NewSimulation is a constructor for the Simulation data type,
// and a pre-processor function for the embedded types.
func NewSimulation(cfgFile string) (*Simulation, error) {
	cfg, cfgErr := loadConfig(cfgFile)
	if cfgErr != nil {
		log.Fatal("Config error reached: ", cfgErr)
		return nil, cfgErr
	}

	startingCash := FloatAmount(cfg.Backtest.StartCashAmt)
	sim := &Simulation{
		config: cfg,
		oms: newOMS(
			startingCash,
			cfg.Simulation.Costmethod,
			cfg.Backtest.IgnoreSecurities,
			cfg.Simulation.OutFmt,
		),
		inputChans: make([]*inputChan, 0),
		// create channels.
		closing:      make(chan struct{}),
		processTicks: make(chan struct{}),
		waitOnInput:  make(chan struct{}),
	}
	return sim, nil
}

// Simulation embeds all data structs necessary for running a backtest of an algorithmic strategy.
type Simulation struct {
	sync.RWMutex
	oms    *OMS
	config *Config
	// Channels
	inputChans   []*inputChan
	closing      chan struct{}
	processTicks chan struct{}
	waitOnInput  chan struct{}
}

// Run acts as the simulation's primary pipeline function; directing everything to where it needs to go.
func (sim *Simulation) Run() {
	log.Println("Starting sim...")
	if sim.oms.strat.algo == nil {
		log.Fatal("Algorithm needs to be implemented by end-user")
	}
	go sim.oms.prfmLog.run()

	log.Println("loading input...")
	go sim.loadInput()

	// <-sim.processTicks // Wait for signal to start processing tick data
	// close(sim.processTicks)

	// go func(tickChan *) {
	// 	for {
	// 		select {
	// 		case tick, ok := <-*tickCh:
	//
	// 			if !ok {
	// 				sim.Lock()
	// 				if len(sim.inputChans) > 1 {
	// 					log.Println(len(sim.inputChans))
	// 					log.Println("re-slicing inputChans")
	// 					sim.inputChans = sim.inputChans[0:]
	// 					sim.Unlock()
	// 					*tickCh = nil
	// 					goto restart
	// 				} else {
	// 					sim.conclude()
	// 					break
	// 				}
	// 			}
	// 			log.Println("Simulating!!!!")
	// 			sim.simulateData(tick)
	// 			// default:
	// 			// 	time.Sleep(1 * time.Millisecond)
	// 		}
	// 	}
	// }()
	<-sim.closing
}

func (sim *Simulation) loadInput() error {
	var filedate time.Time

	sim.RLock()
	simCfg := *sim.config
	sim.RUnlock()
	// log.Println(simCfg.File.Glob)
	fileGlob, globErr := filepath.Glob(simCfg.File.Glob)

	if globErr != nil || len(fileGlob) == 0 {
		log.Fatal("No files matching the glob can be found")
		return globErr
	}

	inputChan := newInputChan()

	go func() {
		for {
			select {
			case record := <-inputChan.recordC:
				sim.prepRecord(inputChan.tickC, record, filedate, simCfg)

			case tick := <-inputChan.tickC:
				sim.oms.processTick(tick)
			}
		}
	}()

	for _, file := range fileGlob {
		log.Println("Loading data from:", file)
		datafile, fileErr := os.Open(file)

		if fileErr != nil {
			log.Fatal("File cannot be loaded")
			panic("File cannot be loaded")
		}
		r := csv.NewReader(datafile)

		if delim, _ := utf8.DecodeRuneInString(simCfg.File.Delim); delim == utf8.RuneError {
			log.Fatal("delim cannot be parsed")
			// return errors.New("File delimiter could not be parsed")
		} else {
			r.Comma = delim
		}
		if simCfg.File.Headers == true {
			r.Read()
		}

		lastUnderscore := strings.LastIndex(file, "_")
		fileDate := file[lastUnderscore+1:]

		lastDate, dateErr := time.Parse(simCfg.File.ExampleDate, fileDate)
		if dateErr != nil {
			log.Fatal("Date cannot be parsed")
		}
		filedate = lastDate

		done := make(chan struct{})
		go func(done chan struct{}) {

			for {
				record, recordErr := r.Read()
				if recordErr != nil {
					if recordErr == io.EOF {
						break
					} else {
						log.Fatal("Error reading record from file")
					}
				}
				inputChan.recordC <- record
			}
			close(inputChan.recordC)
			done <- struct{}{}
		}(done)
		<-done
	}

	go func() { sim.processTicks <- struct{}{} }()

	<-sim.waitOnInput
	return nil
}

// TODO: add logging statements for if things get `hairy`
func (sim *Simulation) prepRecord(outChan chan *Tick, record []string, filedate time.Time, simCfg Config) {
	var loadErr error

	tick := new(Tick)
	tick.Ticker = record[simCfg.File.Columns.Ticker]

	bid, bidErr := strconv.ParseFloat(record[simCfg.File.Columns.Bid], 64)
	if bidErr != nil {
		loadErr = errors.New("Bid Price could not be parsed")
	}
	tick.Bid = FloatAmount(bid)

	bidSize, bidSzErr := strconv.ParseFloat(record[simCfg.File.Columns.BidSize], 64)
	if bidSzErr != nil {
		loadErr = errors.New("Bid Size could not be parsed")
	}
	tick.BidSize = FloatAmount(bidSize)

	ask, askErr := strconv.ParseFloat(record[simCfg.File.Columns.Ask], 64)
	if askErr != nil {
		loadErr = errors.New("Ask Price could not be parsed")
	}
	tick.Ask = FloatAmount(ask)

	askSize, askSzErr := strconv.ParseFloat(record[simCfg.File.Columns.AskSize], 64)
	if askSzErr != nil {
		loadErr = errors.New("Ask Size could not be parsed")
	}
	tick.AskSize = FloatAmount(askSize)

	tickDuration, timeErr := time.ParseDuration(record[simCfg.File.Columns.Timestamp] + simCfg.File.TimestampUnit)

	if timeErr != nil {
		loadErr = timeErr
	}
	tick.Timestamp = filedate.Add(tickDuration)

	if loadErr != nil {
		log.Println(loadErr)
		log.Fatal("record could not be loaded")
	}
	outChan <- tick
}

// TODO:
func (sim *Simulation) conclude() {
	sim.oms.closeHandle()

	sim.close()
}

func (sim *Simulation) simulateData(tick *Tick) {
	if _, exists := sim.oms.strat.ignore[tick.Ticker]; exists {
		return
	}
	go func() { sim.oms.tickChan <- tick }()
}

// close sends a signal to close all channels and exit current processes
func (sim *Simulation) close() {
	sim.closing <- struct{}{}
}

func newInputChan() *inputChan {
	return &inputChan{
		recordC: make(chan []string, 1024),
		tickC:   make(chan *Tick, 1024),
	}
}

type inputChan struct {
	recordC chan []string
	tickC   chan *Tick
}

// func(inChan *inputChan) run() {
// 	go func() {
// 		for inChan.tickC != nil {
// 			select {
// 				case
// 			}
// 		}
// 	}
// }

func (ch *inputChan) deconstruct() {
	close(ch.recordC)
	close(ch.tickC)
}
