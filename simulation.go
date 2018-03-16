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
	"time"
)

// LoadAlgorithm ensures that an Algorithm interface is implemented in the Simulation pipeline to be used by other functions.
func LoadAlgorithm(sim *Simulation, algo Algorithm) bool {
	sim.btEngine.OMS.strategy.algo = algo
	return true
}

// NewSimulation is a constructor for the Simulation data type,
// and a pre-processor function for the embedded types.
func NewSimulation(cfgFile string) (*Simulation, error) {
	cfg, cfgErr := loadConfig(cfgFile)
	if cfgErr != nil {
		return nil, cfgErr
	}
	startingCash := FloatAmount(cfg.Backtest.StartCashAmt)
	sim := &Simulation{
		// REVIEW: may want to swap btEngine with OMS...or even embed OMS instead.
		btEngine: newBacktestEngine(
			startingCash,
			cfg.Simulation.Costmethod,
			cfg.Backtest.IgnoreSecurities,
		),
		inputChans: make([]*inputChan, 0),
		// create channels.
		closing:    make(chan struct{}),
		loadChan:   make(chan bool, 1),
		startInput: make(chan bool, 1),
	}
	return sim, nil
}

// Simulation embeds all data structs necessary for running a backtest of an algorithmic strategy.
type Simulation struct {
	btEngine *BacktestEngine
	config   *Config
	// Channels
	closing    chan struct{}
	inputChans []*inputChan
	loadChan   chan bool
	startInput chan bool
}

// run acts as the simulation's primary pipeline function; directing everything to where it needs to go.
func (sim *Simulation) run() {

	if sim.btEngine.OMS.strategy.algo == nil {
		panic("Algorithm needs to be implemented by end-user")
	}
	sim.startInput <- true

	go func() {
		for {
			select {
			case <-sim.startInput:
				sim.loadInput()
				sim.startInput = nil

			case <-sim.loadChan:
				if len(sim.inputChans) > 1 {
					sim.popFly()
				} else {
					sim.conclude()
				}
				// default:
				// 	time.Sleep(1 * time.Millisecond)
			}
		}
	}()
	<-sim.closing
}

// TODO:
func (sim *Simulation) conclude() {
	sim.btEngine.OMS.closeHandle()

	sim.close()
}

// TODO: rename later to make more sense
func (sim *Simulation) popFly() {
	done := make(chan struct{})

	sim.inputChans[0].deconstruct()

	sim.inputChans = sim.inputChans[0:]
	inChan := sim.inputChans[0]

	go func() {
		for tick := range inChan.tickC {
			go func(t *Tick) { sim.simulateData(t) }(tick)
		}
		done <- struct{}{}
	}()
	<-done
}

func (sim *Simulation) simulateData(tick *Tick) {
	if _, exists := sim.btEngine.OMS.strategy.ignore[tick.Ticker]; exists {
		return
	}
	go func() { sim.btEngine.OMS.tickChan <- tick }()
}

func (sim *Simulation) loadInput() error {
	fileGlob, globErr := filepath.Glob(sim.config.File.Glob)
	if globErr != nil {
		return globErr
	}
	// add extra bucket for sentinel value
	sim.inputChans = make([]*inputChan, len(fileGlob)+1)
	sim.inputChans[0] = new(inputChan)

	go func() {
		for i, file := range fileGlob {
			sim.inputChans[i+1] = newInputChan()
			sim.loadData(sim.inputChans[i+1], file)
		}
	}()
	sim.popFly()

	return nil
}

func (sim *Simulation) loadData(inChan *inputChan, file string) (err error) {
	quit := make(chan struct{})

	datafile, fileErr := os.Open(file)
	defer datafile.Close()
	if fileErr != nil {
		log.Fatal("File cannot be loaded")
		err = fileErr
		return
	}

	log.Printf("Started loading data from file: %s", file)
	r := csv.NewReader(datafile)
	r.Comma = sim.config.File.Delim

	go func() {
		for {
			record, recordErr := r.Read()
			if recordErr == io.EOF {
				break
			}
			inChan.recordC <- record
		}
		close(inChan.recordC)
	}()

	lastDate, dateErr := time.Parse(
		sim.config.File.ExampleDate, strings.TrimSuffix(file, filepath.Ext(file)))
	if dateErr != nil {
		err = dateErr
		return
	}
	go sim.loadTicks(quit, inChan, lastDate)

	// TODO:
	// if tickErr := <-quit; tickErr != nil {
	// 	return tickErr
	// }
	<-quit
	return
}

// TODO: add logging statements for if things get `hairy`
func (sim *Simulation) loadTicks(quit chan<- struct{}, inChan *inputChan, date time.Time) error {
	for record := range inChan.recordC {
		var tick *Tick
		tick.Ticker = record[sim.config.File.Columns.Ticker]

		bid, bidErr := strconv.ParseFloat(record[sim.config.File.Columns.Bid], 64)
		if bidErr != nil {
			return errors.New("Bid Price could not be parsed")
		}
		tick.Bid = FloatAmount(bid)

		bidSize, bidSzErr := strconv.ParseFloat(record[sim.config.File.Columns.BidSize], 64)
		if bidSzErr != nil {
			return errors.New("Bid Size could not be parsed")
		}
		tick.BidSize = FloatAmount(bidSize)

		ask, askErr := strconv.ParseFloat(record[sim.config.File.Columns.Ask], 64)
		if askErr != nil {
			return errors.New("Ask Price could not be parsed")
		}
		tick.Ask = FloatAmount(ask)

		askSize, askSzErr := strconv.ParseFloat(record[sim.config.File.Columns.AskSize], 64)
		if askSzErr != nil {
			return errors.New("Ask Size could not be parsed")
		}
		tick.AskSize = FloatAmount(askSize)

		tickDuration, timeErr := time.ParseDuration(record[sim.config.File.Columns.Timestamp] + sim.config.timeUnit())
		if timeErr != nil {
			return timeErr
		}
		tick.Timestamp = date.Add(tickDuration)
		inChan.tickC <- tick
	}
	close(inChan.tickC)

	quit <- struct{}{}

	sim.loadChan <- true
	return nil
}

// close sends a signal to close all channels and exit current processes
func (sim *Simulation) close() {
	sim.closing <- struct{}{}
}

func newInputChan() *inputChan {
	return &inputChan{
		recordC: make(chan []string),
		tickC:   make(chan *Tick),
	}
}

type inputChan struct {
	recordC chan []string
	tickC   chan *Tick
}

// REVIEW: do we need a for-select statement for an inputChan?

func (ch *inputChan) close() {
	close(ch.recordC)
	close(ch.tickC)
	return
}

func (ch *inputChan) deconstruct() {
	ch.close()
	ch.recordC = nil
	ch.tickC = nil
	ch = nil
	return
}
