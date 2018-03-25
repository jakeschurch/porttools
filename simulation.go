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
		tickMeta:     make(chan chan *Tick),
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
	tickMeta     chan chan *Tick
	processTicks chan struct{}
	waitOnInput  chan struct{}
}

// Run acts as the simulation's primary pipeline function; directing everything to where it needs to go.
func (sim *Simulation) Run() {
	log.Println("Starting sim...")
	if sim.oms.strat.algo == nil {
		log.Fatal("Algorithm needs to be implemented by end-user")
	}
	go sim.oms.handle()
	go sim.oms.prfmLog.run()

	log.Println("loading input...")
	go sim.loadInput()

	<-sim.processTicks // Wait for signal to start processing tick data
	close(sim.processTicks)

	go func() {
		for {
			select {
			case tickCh, ok := <-sim.tickMeta:

				if !ok {
					break
				}
				log.Println("Simulating!!!!")
				sim.oms.processTick(tickCh)
				// default:
				// 	time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	for index := range sim.inputChans {
		if index > 0 {
			sim.tickMeta <- sim.inputChans[index].tickC
		}
	}

	<-sim.closing
}

func (sim *Simulation) loadInput() error {
	sim.RLock()
	simCfg := *sim.config
	sim.RUnlock()
	// log.Println(simCfg.File.Glob)
	fileGlob, globErr := filepath.Glob(simCfg.File.Glob)

	if globErr != nil || len(fileGlob) == 0 {
		log.Fatal("No files matching the glob can be found")
		return globErr
	}

	// add extra bucket for sentinel value
	sim.inputChans = make([]*inputChan, len(fileGlob)+1)
	sim.inputChans[0] = newInputChan()
	close(sim.inputChans[0].tickC)

	go func() { sim.processTicks <- struct{}{} }()

	sim.RLock()
	for i, file := range fileGlob {
		inChan := sim.inputChans[i+1]
		inChan = newInputChan()
		go func(index int, file string) { sim.loadData(inChan, file, simCfg) }(i, file)
	}
	sim.RUnlock()

	<-sim.waitOnInput
	log.Println("done waiting")
	return nil
}

func (sim *Simulation) loadData(inChan *inputChan, file string, simCfg Config) (err error) {
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

	go func() {
		for {
			record, recordErr := r.Read()
			if recordErr != nil {
				if recordErr == io.EOF {
					break
				} else {
					log.Fatal("Error reading record from file")
				}
			}
			inChan.recordC <- record
		}
		close(inChan.recordC)
	}()

	lastUnderscore := strings.LastIndex(file, "_")
	fileDate := file[lastUnderscore+1:]

	lastDate, dateErr := time.Parse(simCfg.File.ExampleDate, fileDate)
	if dateErr != nil {
		log.Fatal("Date cannot be parsed")
	}

	log.Println("Loading ticks")
	quit := make(chan struct{})
	go sim.loadTicks(quit, inChan, lastDate, simCfg)

	// TODO:
	// if tickErr := <-quit; tickErr != nil {
	// 	return tickErr
	// }
	<-quit
	log.Println("quitting read")
	datafile.Close()
	log.Println("Done reading data from", file)
	return
}

// TODO: add logging statements for if things get `hairy`
func (sim *Simulation) loadTicks(quit chan<- struct{}, inChan *inputChan, date time.Time, simCfg Config) {
	var loadErr error
	for record := range inChan.recordC {

		log.Println("loading tick...")
		tick := new(Tick)
		tick.Ticker = record[simCfg.File.Columns.Ticker]
		log.Println(tick.Ticker)

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
		tick.Timestamp = date.Add(tickDuration)

		if loadErr != nil {
			log.Println(loadErr)
			log.Fatal("record could not be loaded")
		}
		inChan.tickC <- tick
	}
	close(inChan.tickC)
	quit <- struct{}{}
}

// TODO:
func (sim *Simulation) conclude() {
	sim.oms.closeHandle()

	sim.close()
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
