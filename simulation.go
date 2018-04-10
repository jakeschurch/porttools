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

	"github.com/jakeschurch/porttools/collection/benchmark"
	"github.com/jakeschurch/porttools/collection/portfolio"
	"github.com/jakeschurch/porttools/config"
	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/output"
	"github.com/jakeschurch/porttools/utils"
)

var (
	Oms         *OMS
	Port        *portfolio.Portfolio
	positionLog *output.PositionLog
	index       *benchmark.Index
	strategy    Strategy
	simConfig   config.Config
	costMethod  utils.CostMethod

	// ErrInvalidFileGlob indiciates that no files could be found from given glob
	ErrInvalidFileGlob = errors.New("No files could be found from file glob")

	// ErrInvalidFileDelim is thrown when file delimiter is not able to be parsed
	ErrInvalidFileDelim = errors.New("File delimiter could not be parsed")
)

func init() {
	Oms = NewOMS()
	Port = portfolio.New()
	positionLog = output.NewPositionLog()
	index = benchmark.NewIndex()
}

// NewSimulation is a constructor for the Simulation data type,
// and a pre-processor function for the embedded types.
func NewSimulation(file string) (*Simulation, error) {
	simConfig, simConfigErr := config.Load(file)
	if simConfigErr != nil {
		log.Fatal("Config error reached: ", simConfigErr)
		return nil, simConfigErr
	}
	costMethod = simConfig.Simulation.Costmethod
	Port.UpdateCash(utils.FloatAmount(simConfig.Backtest.StartCashAmt))
	sim := &Simulation{
		// Channels
		processChan: make(chan *instrument.Tick),
		tickChan:    make(chan *instrument.Tick),
		errChan:     make(chan error),
	}
	log.Println("Created sim")
	return sim, nil
}

// Simulation embeds all data structs necessary for running a backtest of an algorithmic strategy.
type Simulation struct {
	mu          sync.RWMutex
	processChan chan *instrument.Tick
	tickChan    chan *instrument.Tick
	errChan     chan error
}

// Run acts as the simulation's primary pipeline function; directing everything to where it needs to go.
func (sim *Simulation) Run() error {
	log.Println("Starting sim...")
	if strategy.Algorithm == nil {
		log.Fatal("Algorithm needs to be implemented by end-user")
	}

	done := make(chan struct{})
	go func() {
		cachedTicks := make([]*instrument.Tick, 0)

		for sim.tickChan != nil {
			tick, ok := <-sim.tickChan
			if !ok {
				if tick != nil {
					cachedTicks = append(cachedTicks, tick)
					// assuming rest of ticks are nil, loop over and process the remaining ticks
				} else {
					for i := range cachedTicks {
						sim.process(cachedTicks[i])
					}
					break
				}
			}
			if tick != nil {
				sim.process(tick)
			}
		}
		close(done)
	}()

	log.Println("loading input...")
	fileName, fileDate := fileInfo()

	// DO NOT REVIEW
	colConfig := colConfig{tick: simConfig.File.Columns.Ticker,
		bid:      simConfig.File.Columns.Bid,
		bidSz:    simConfig.File.Columns.BidSize,
		ask:      simConfig.File.Columns.Ask,
		askSz:    simConfig.File.Columns.AskSize,
		filedate: fileDate,
		timeUnit: simConfig.File.TimestampUnit,
	}

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	worker := newWorker(colConfig)
	go worker.run(sim.tickChan, file)

	<-done
	log.Println(positionLog.ClosedPositions)
	output.GetResults(output.CSV, positionLog.ClosedPositions, index.Holdings)

	return nil
}

func fileInfo() (string, time.Time) {
	fileGlob, err := filepath.Glob(simConfig.File.Glob)
	if err != nil || len(fileGlob) == 0 {
		// return ErrInvalidFileGlob
	}
	file := fileGlob[0]
	lastUnderscore := strings.LastIndex(file, "_")
	fileDate := file[lastUnderscore+1:]

	lastDate, dateErr := time.Parse(simConfig.File.ExampleDate, fileDate)
	if dateErr != nil {
		log.Fatal("Date cannot be parsed")
	}
	filedate := lastDate
	return file, filedate
}

// Process simulates tick data going through our simulation pipeline
func (sim *Simulation) process(t *instrument.Tick) error {

	Oms.Query(*t)

	Port.Update(*t.Quote)

	index.Update(*t.Quote)

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

func newWorker(cols colConfig) *worker {
	worker := &worker{
		colCfg: cols,
	}
	return worker
}

func (worker *worker) run(outChan chan<- *instrument.Tick, r io.ReadSeeker) {
	var lineCount int
	done := make(chan struct{}, 2)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lineCount++
	}
	r.Seek(0, 0)

	worker.dataChan = make(chan []string, lineCount)
	go worker.send(outChan, done)
	go worker.produce(done, r)

	<-done
	<-done
}

func (worker *worker) send(outChan chan<- *instrument.Tick, done chan struct{}) {
	for {
		data, ok := <-worker.dataChan
		if !ok {
			if len(worker.dataChan) == 0 {
				close(outChan)
				break
			}
		}
		tick, err := worker.consume(data)
		if tick != nil && err == nil {
			outChan <- tick
		}
	}
	done <- struct{}{}
}

// 3 by 2 feet
func (worker *worker) produce(done chan struct{}, r io.ReadSeeker) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	scanner.Scan() // for headers...
	for scanner.Scan() {
		line := scanner.Text()

		// Check to see if error has been thrown or
		if err := scanner.Err(); err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalln(err)
			}
		}
		record := strings.Split(line, "|")
		if len(record) > 4 {
			worker.dataChan <- record
		}
	}
	close(worker.dataChan)
	log.Println("done reading from file")
	done <- struct{}{}
}

func (worker *worker) consume(record []string) (*instrument.Tick, error) {
	var loadErr, parseErr error
	var tick *instrument.Tick

	tick = new(instrument.Tick)
	tick.SetTicker(record[worker.colCfg.tick])

	bid, bidErr := strconv.ParseFloat(record[worker.colCfg.bid], 64)
	if bid == 0 {
		return tick, errors.New("bid Price could not be parsed")
	}
	if bidErr != nil {
		loadErr = errors.New("bid Price could not be parsed")
	}
	tick.Bid = utils.FloatAmount(bid)

	bidSz, bidSzErr := strconv.ParseFloat(record[worker.colCfg.bidSz], 64)
	if bidSzErr != nil {
		loadErr = errors.New("bid Size could not be parsed")
	}
	tick.BidSize = utils.Amount(bidSz)

	ask, askErr := strconv.ParseFloat(record[worker.colCfg.ask], 64)
	if ask == 0 {
		return nil, askErr
	}
	if askErr != nil {
		loadErr = errors.New("ask Price could not be parsed")
	}
	tick.Ask = utils.FloatAmount(ask)

	askSz, askSzErr := strconv.ParseFloat(record[worker.colCfg.askSz], 64)
	if askSzErr != nil {
		loadErr = errors.New("ask Size could not be parsed")
	}
	tick.AskSize = utils.Amount(askSz)

	tickDuration, timeErr := time.ParseDuration(record[worker.colCfg.tStamp] + worker.colCfg.timeUnit)

	if timeErr != nil {
		loadErr = timeErr
	}
	tick.Timestamp = worker.colCfg.filedate.Add(tickDuration)

	if parseErr != nil {
		return tick, parseErr
	}
	if loadErr != nil {
		log.Fatal("record could not be loaded")
	}
	return tick, nil
}
