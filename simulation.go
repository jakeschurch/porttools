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

// NewSimulation is a constructor for the Simulation data type,
// and a pre-processor function for the embedded types.
func NewSimulation(cfgFile string) (*Simulation, error) {
	cfg, cfgErr := loadConfig(cfgFile)
	if cfgErr != nil {
		return nil, cfgErr
	}
	startingCash := FloatAmount(cfg.Backtest.StartCashAmt)
	sim := &Simulation{
		btEngine:   newBacktestEngine(startingCash, cfg.Backtest.IgnoreSecurities),
		closing:    make(chan chan error),
		inputChans: make([]*inputChan, 0),
	}
	return sim, nil
}

// LoadAlgorithm ensures that an Algorithm interface is implemented in the Simulation pipeline to be used by other functions.
func LoadAlgorithm(sim *Simulation, algo Algorithm) bool {
	sim.btEngine.OMS.strategy.Algorithm = algo
	return true
}

// Simulation embeds all data structs necessary for running a backtest of an algorithmic strategy.
type Simulation struct {
	btEngine *BacktestEngine
	config   *Config // TODO: remove config
	// Channels
	closing    chan chan error
	inputChans []*inputChan
}

func (sim *Simulation) run() {
	done := make(chan struct{})
	var err error
	// TODO: err checking/initialize, etc.
	go func() {
		for {
			select {
			case errc := <-sim.closing:
				errc <- err
				for _, c := range sim.inputChans {
					c.deconstruct()
				}
				done <- struct{}{}
				break
			}
		}
	}()
	<-done
}

// TODO: rename later to make more sense
func (sim *Simulation) popFly() {
	done := make(chan struct{})

	sim.inputChans[0].deconstruct()

	sim.inputChans = sim.inputChans[0:]
	inChan := sim.inputChans[0]
	go func() {
		for tick := range inChan.tickC {
			go sim.simulateData(tick)
		}
		done <- struct{}{}
	}()

	<-done
	if len(sim.inputChans) > 1 {
		sim.popFly()
	} else {
		return
	}
}

func (sim *Simulation) simulateData(tick *Tick) {
	if _, exists := sim.btEngine.OMS.strategy.ignoreTickers[tick.Ticker]; exists {
		return
	}
	go sim.btEngine.OMS.Portfolio.updatePositions(tick)
	go sim.btEngine.Benchmark.updateSecurities(tick)
}

func (sim *Simulation) loadInput() error {
	fileGlob, globErr := filepath.Glob(sim.config.File.Glob)
	if globErr != nil {
		return globErr
	}
	// extra bucket for sentinel value
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

func (sim *Simulation) loadData(inChan *inputChan, file string) error {
	quit := make(chan struct{})

	datafile, fileErr := os.Open(file)
	defer datafile.Close()
	if fileErr != nil {
		log.Fatal("File cannot be loaded")
		return fileErr
	}
	r := csv.NewReader(datafile)
	r.Comma = sim.config.File.Delim

	go func() {
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			inChan.recordC <- record
		}
		close(inChan.recordC)
	}()

	lastDate, dateErr := time.Parse(
		sim.config.File.ExampleDate, strings.TrimSuffix(file, filepath.Ext(file)))
	if dateErr != nil {
		return dateErr
	}
	go sim.loadTicks(quit, inChan, lastDate)

	<-quit
	return nil
}

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
	quit <- struct{}{}
	return nil
}

func (sim *Simulation) close() error {
	errc := make(chan error)
	sim.closing <- errc
	return <-errc
}

func newInputChan() *inputChan {
	inputChan := new(inputChan)
	inputChan.recordC = make(chan []string)
	inputChan.tickC = make(chan *Tick)
	return inputChan
}

type inputChan struct {
	recordC chan []string
	tickC   chan *Tick
}

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
