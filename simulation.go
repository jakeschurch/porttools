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

var (
	ErrInvalidFileGlob  = errors.New("No files could be found from file glob")
	ErrInvalidFileDelim = errors.New("File delimiter could not be parsed")
)

// LoadAlgorithm ensures that an Algorithm interface is implemented in the Simulation pipeline to be used by other functions.
func (sim *Simulation) LoadAlgorithm(algo Algorithm) bool {
	sim.strategy = newStrategy(algo, []string{})
	return true
}

// NewSimulation is a constructor for the Simulation data type,
// and a pre-processor function for the embedded types.
func NewSimulation(algo Algorithm, cfgFile string) (*Simulation, error) {
	cfg, cfgErr := loadConfig(cfgFile)
	if cfgErr != nil {
		log.Fatal("Config error reached: ", cfgErr)
		return nil, cfgErr
	}

	startingCash := FloatAmount(cfg.Backtest.StartCashAmt)
	sim := &Simulation{
		config:    *cfg,
		oms:       newOMS(),
		port:      NewPortfolio(startingCash),
		prfmLog:   newPrfmLog(),
		benchmark: NewIndex(),
		strategy:  newStrategy(algo, []string{}),
		// Channels
		processChan: make(chan *Tick),
		tickChan:    make(chan *Tick),
		errChan:     make(chan error),
	}
	return sim, nil
}

// Simulation embeds all data structs necessary for running a backtest of an algorithmic strategy.
type Simulation struct {
	sync.RWMutex
	oms         *OMS
	config      Config
	port        *Portfolio
	prfmLog     *PrfmLog
	benchmark   *Index
	strategy    strategy
	processChan chan *Tick
	tickChan    chan *Tick
	errChan     chan error
}

// Run acts as the simulation's primary pipeline function; directing everything to where it needs to go.
func (sim *Simulation) Run() error {
	// recordChan := make(chan []string)

	log.Println("Starting sim...")
	if sim.strategy.algorithm == nil {
		log.Fatal("Algorithm needs to be implemented by end-user")
	}

	log.Println("loading input...")
	_, filedate := fileInfo(sim.config)
	go sim.setup(sim.tickChan, filedate, sim.config)

	// counter := 0
	done := make(chan struct{})

	go func() {
		amtLeft := 0
		for sim.tickChan != nil {
			select {
			// case record, ok := <-recordChan:
			// 	if !ok {
			// 		recordChan = nil
			// 		close()
			// 		continue
			// 	}
			// 	counter++
			// 	log.Printf("\rTotal of %d ticks processed", counter)
			// 	if record != nil {
			// 		go sim.prepRecord(record, filedate, sim.config)
			// 	}

			case tick, ok := <-sim.tickChan:
				amtLeft++
				if !ok {
					sim.tickChan = nil
					close(done)
				}
				if tick != nil {
					sim.process(tick)
				}

				// case tickToProcess, ok := <-sim.processChan:
				// 	amtLeft--
				// 	if !ok {
				// 		close(done)
				// 		sim.processChan = nil
				// 		return
				// 	}
				// 	if tickToProcess != nil {
				// 		sim.errChan <- sim.process(tickToProcess)
				// 	}
			}
		}
	}()
	<-done
	log.Println(len(sim.prfmLog.closedPositions))
	log.Println(len(sim.prfmLog.closedOrders))
	getResults(sim.prfmLog.closedPositions, sim.benchmark.Securities, sim.config.Simulation.OutFmt)

	return nil
}

func fileInfo(cfg Config) (string, time.Time) {
	fileGlob, err := filepath.Glob(cfg.File.Glob)
	if err != nil || len(fileGlob) == 0 {
		// return ErrInvalidFileGlob
	}
	file := fileGlob[0]
	lastUnderscore := strings.LastIndex(file, "_")
	fileDate := file[lastUnderscore+1:]

	lastDate, dateErr := time.Parse(cfg.File.ExampleDate, fileDate)
	if dateErr != nil {
		log.Fatal("Date cannot be parsed")
	}
	filedate := lastDate
	return file, filedate
}

func (sim *Simulation) setup(tickChan chan<- *Tick, filedate time.Time, cfg Config) {
	// Check if valid file path/s
	fileGlob, err := filepath.Glob(cfg.File.Glob)
	if err != nil || len(fileGlob) == 0 {
		log.Fatal(ErrInvalidFileGlob)
	}
	datafile := fileGlob[0]
	file, err := os.Open(datafile)
	if err != nil {
		log.Fatal(err)
	}
	r := csv.NewReader(file)

	delim, _ := utf8.DecodeRuneInString(cfg.File.Delim)
	// if delim == utf8.RuneError {
	// 	log.Println("uh oh")
	// return ErrInvalidFileDelim
	// }
	r.Comma = delim
	if cfg.File.Headers == true {
		r.Read()
	}
	for {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				log.Println("EOF")
				break
			}
		}
		tickChan <- sim.prepRecord(record, filedate, cfg)
	}
	log.Println("closing tickchan")
	close(sim.tickChan)
}

func (sim *Simulation) loadRecords(outChan chan<- []string, records [][]string) {
	for _, record := range records[1:] {
		outChan <- record
	}
	close(outChan)
}

func (sim *Simulation) prepRecord(record []string, filedate time.Time, simCfg Config) *Tick {
	var loadErr, parseErr error
	var tick *Tick

	tick = new(Tick)
	tick.Ticker = record[simCfg.File.Columns.Ticker]

	bid, bidErr := strconv.ParseFloat(record[simCfg.File.Columns.Bid], 64)
	if bid == 0 {
		return nil
	}
	if bidErr != nil {
		loadErr = errors.New("Bid Price could not be parsed")
	}
	tick.Bid = FloatAmount(bid)

	bidSize, bidSzErr := strconv.ParseFloat(record[simCfg.File.Columns.BidSize], 64)
	if bidSzErr != nil {
		loadErr = errors.New("Bid Size could not be parsed")
	}
	tick.BidSize = Amount(bidSize)

	ask, askErr := strconv.ParseFloat(record[simCfg.File.Columns.Ask], 64)
	if ask == 0 {
		return nil
	}
	if askErr != nil {
		loadErr = errors.New("Ask Price could not be parsed")
	}
	tick.Ask = FloatAmount(ask)

	askSize, askSzErr := strconv.ParseFloat(record[simCfg.File.Columns.AskSize], 64)
	if askSzErr != nil {
		loadErr = errors.New("Ask Size could not be parsed")
	}
	tick.AskSize = Amount(askSize)

	tickDuration, timeErr := time.ParseDuration(record[simCfg.File.Columns.Timestamp] + simCfg.File.TimestampUnit)

	if timeErr != nil {
		loadErr = timeErr
	}
	tick.Timestamp = filedate.Add(tickDuration)

	if parseErr != nil {
		return nil
	}
	if loadErr != nil {
		log.Println(loadErr)
		log.Fatal("record could not be loaded")
	}
	return tick
}

// Process simulates tick data going through our simulation pipeline
func (sim *Simulation) process(tick *Tick) error {

	// Update benchmark metrics
	if err := sim.benchmark.UpdateMetrics(*tick); err != nil {
		if err == ErrNoSecurityExists {
			sim.benchmark.AddNew(*tick)
		} else {
			return err
		}
	}
	// Add entry order if it meets valid order logic
	if newOrder, err := sim.strategy.checkEntryLogic(sim.port, *tick); err == nil {
		txAmount, err := sim.oms.addOrder(newOrder)
		if err != nil {
			return err
		}
		// create new position from order
		newPos := newOrder.toPosition()
		// add new position (holding) and change in cash from order to portfolio
		if err := sim.port.AddHolding(newPos, txAmount); err != nil {
			if err == ErrNoSliceExists {
				sim.port.AddNew(newPos, txAmount)
			} else {
				return err
			}

			log.Println(sim.port)

		}
	}

	// Check if open order with same ticker exists
	if matchedOrders, err := sim.oms.existsInOrders(tick.Ticker); err != ErrEmptySlice {
		// if openOrder with matching ticker exists, it means we are holding a position in our portfolio
		// - so update portfolio metrics

		if err := sim.port.UpdateMetrics(*tick); err != nil {
			return err
		}

		for _, matchedOrder := range matchedOrders {
			if newClosedOrder, err := sim.strategy.checkExitLogic(sim.port, matchedOrder, *tick); err == nil {

				sim.port.Lock()
				txAmount, closedPositions, deleteSlice, err := sim.oms.TransactSell(newClosedOrder,
					sim.config.Simulation.Costmethod,
					sim.port.active[tick.Ticker])
				sim.port.Unlock()

				if err != nil {
					return err
				}
				// Update held Cash amount in portfolio
				sim.port.UpdateCash(txAmount)

				// Delete holding slice from portfolio active holdings map if now empty
				if deleteSlice {
					sim.port.Lock()
					delete(sim.port.active, tick.Ticker)
					sim.port.Unlock()
				}

				// Add closed positions (holdings) to performance log
				for _, closedPos := range closedPositions {
					sim.prfmLog.addPosition(closedPos)
				}

				// Add closed order to performance log
				if err := sim.prfmLog.addOrder(newClosedOrder); err != nil {
					return err
				}
			}
		}
	}
	log.Println("all done")
	return nil
}
