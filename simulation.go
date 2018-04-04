package porttools

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	// ErrInvalidFileGlob indiciates that no files could be found from given glob
	ErrInvalidFileGlob = errors.New("No files could be found from file glob")

	// ErrInvalidFileDelim is thrown when file delimiter is not able to be parsed
	ErrInvalidFileDelim = errors.New("File delimiter could not be parsed")
)

// LoadAlgorithm ensures that an Algorithm interface is implemented in the Simulation pipeline to be used by other functions.
func (sim *Simulation) LoadAlgorithm(algo Algorithm) bool {
	sim.strategy = newStrategy(algo, make([]string, 0))
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
		strategy:  newStrategy(algo, make([]string, 0)),
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
	log.Println("Starting sim...")
	if sim.strategy.algorithm == nil {
		log.Fatal("Algorithm needs to be implemented by end-user")
	}

	done := make(chan struct{})
	go func() {
		for sim.tickChan != nil {
			tick, ok := <-sim.tickChan
			if !ok {
				close(done)
			}
			if tick != nil {
				sim.process(tick)
			}
		}
	}()

	log.Println("loading input...")
	fileName, fileDate := fileInfo(sim.config)

	// DO NOT REVIEW
	colConfig := colConfig{tick: sim.config.File.Columns.Ticker,
		bid:      sim.config.File.Columns.Bid,
		bidSz:    sim.config.File.Columns.BidSize,
		ask:      sim.config.File.Columns.Ask,
		askSz:    sim.config.File.Columns.AskSize,
		filedate: fileDate,
		timeUnit: sim.config.File.TimestampUnit,
	}

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	worker := newWorker(sim.tickChan, colConfig)
	go worker.run(sim.tickChan, file)

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

// func (sim *Simulation) setup(tickChan chan<- *Tick, filedate time.Time, cfg Config) {
// 	// Check if valid file path/s
// 	fileGlob, err := filepath.Glob(cfg.File.Glob)
// 	if err != nil || len(fileGlob) == 0 {
// 		log.Fatal(ErrInvalidFileGlob)
// 	}
// 	datafile := fileGlob[0]
// 	file, err := os.Open(datafile)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	r := csv.NewReader(file)

// 	delim, _ := utf8.DecodeRuneInString(cfg.File.Delim)
// 	r.Comma = delim
// 	if cfg.File.Headers == true {
// 		r.Read()
// 	}
// 	for {
// 		record, err := r.Read()
// 		if err != nil {
// 			if err == io.EOF {
// 				log.Println("EOF")
// 				break
// 			}
// 		}
// 		tickChan <- sim.prepRecord(record, filedate, cfg)
// 	}
// 	log.Println("closing tickchan")
// 	close(sim.tickChan)
// }

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

				sim.port.RLock()
				holdingSlice := sim.port.active[tick.Ticker]
				sim.port.RUnlock()

				txAmount, closedPositions, deleteSlice, err := sim.oms.TransactSell(newClosedOrder,
					sim.config.Simulation.Costmethod,
					holdingSlice)

				if err != nil {
					log.Println("debug")
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
	return nil
}

type colConfig struct {
	tick, bid, bidSz, ask, askSz, tStamp uint8
	filedate                             time.Time
	timeUnit                             string
}

type worker struct {
	dataChan chan []string
	colCfg   colConfig
}

func newWorker(outChan chan<- *Tick, cols colConfig) *worker {
	worker := &worker{
		colCfg: cols,
	}
	return worker
}

func (worker *worker) run(outChan chan<- *Tick, r io.ReadSeeker) {
	var lineCount int
	done := make(chan struct{}, 2)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lineCount++
	}
	r.Seek(0, 0)

	worker.dataChan = make(chan []string, 10000)
	go worker.send(outChan, done)
	go worker.produce(done, r)

	<-done
	<-done
}

func (worker *worker) send(outChan chan<- *Tick, done chan struct{}) {
	for {
		data, ok := <-worker.dataChan
		log.Println(data)
		if !ok {
			if len(worker.dataChan) == 0 {
				break
			}
		}
		if tick := worker.consume(data); tick != nil {
			outChan <- tick
		}
	}
	log.Println("Out")
	done <- struct{}{}
}

func (worker *worker) produce(done chan struct{}, r io.ReadSeeker) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	log.Println(scanner.Scan())
	for scanner.Scan() {
		line := scanner.Text()

		// Check to see if error has been thrown or
		if err := scanner.Err(); err != nil {
			if err == io.EOF {
				log.Println("done here")
				break
			} else {
				log.Println(err)
			}
		}
		record := strings.Split(line, "|")
		log.Println(record)
		if len(record) > 4 {
			worker.dataChan <- record
		}
	}
	close(worker.dataChan)
	log.Println("done")
	done <- struct{}{}
}

func (worker *worker) consume(record []string) *Tick {
	var loadErr, parseErr error
	var tick *Tick

	tick = new(Tick)
	log.Println(record)
	tick.Ticker = record[worker.colCfg.tick]

	bid, bidErr := strconv.ParseFloat(record[worker.colCfg.bid], 64)
	if bid == 0 {
		return nil
	}
	if bidErr != nil {
		loadErr = errors.New("bid Price could not be parsed")
	}
	tick.Bid = FloatAmount(bid)

	bidSz, bidSzErr := strconv.ParseFloat(record[worker.colCfg.bidSz], 64)
	if bidSzErr != nil {
		loadErr = errors.New("bid Size could not be parsed")
	}
	tick.BidSize = Amount(bidSz)

	ask, askErr := strconv.ParseFloat(record[worker.colCfg.ask], 64)
	if ask == 0 {
		return nil
	}
	if askErr != nil {
		loadErr = errors.New("ask Price could not be parsed")
	}
	tick.Ask = FloatAmount(ask)

	askSz, askSzErr := strconv.ParseFloat(record[worker.colCfg.askSz], 64)
	if askSzErr != nil {
		loadErr = errors.New("ask Size could not be parsed")
	}
	tick.AskSize = Amount(askSz)

	tickDuration, timeErr := time.ParseDuration(record[worker.colCfg.tStamp] + worker.colCfg.timeUnit)

	if timeErr != nil {
		loadErr = timeErr
	}
	tick.Timestamp = worker.colCfg.filedate.Add(tickDuration)

	if parseErr != nil {
		return nil
	}
	if loadErr != nil {
		log.Println(loadErr)
		log.Fatal("record could not be loaded")
	}
	return tick
}
