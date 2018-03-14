package porttools

type BacktestEngine struct {
	Portfolio *Portfolio
	Benchmark *Index
	Log       *PerformanceLog
	Strategy  *Strategy
}

// TODO: Finish method signature
type algorithm interface {
	EntryLogic() bool
	ExitLogic() bool
}

// Algorithm ... TODO
type Algorithm struct{}

// Strategy ... TODO
type Strategy struct {
	algos []Algorithm
}

// PerformanceLog conducts performance analysis.
type PerformanceLog struct {
	Closed    PositionSlice
	orders    Queue
	benchmark *Index
}

/* TODO:
	- Restricted tickers
 	- MaxOrderSize
	- MaxPositionSize
	- LongOnly/ShortOnly
	- AssetDateBounds
*/

// TODO: ExecuteStrategy method

//
// import (
// 	"encoding/csv"
// 	"io"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"time"
// )
//
// func notMain() {
//
// 	// example architecture
// 	// 	stream <- Tick
// 	// 	map[Tick.Ticker] <- stream
// 	// 	process <- map
// 	// ticksIn := make(map[string](chan Tick))
// 	//
// 	// doneProcessing := false
// 	// for doneProcessing == false {
// 	// 	// Loop over file
// 	// 	// store each record in a Tick reference and pass it to ticksIn
// 	// 	exampleTick := Tick{Ticker: "AAPL", bid: 101.14, Volume: 20000.00, Datetime: time.Now()}
//
// 	// go func() {
// 	// 	if recordChannel, exists := ticksIn[exampleTick.Ticker]; !exists {
// 	// 		ticksIn[exampleTick.Ticker] = make(chan Tick)
// 	// 	}
// 	// 	ticksIn[exampleTick.Ticker] <- exampleTick
// 	// }()
// 	// option to throw out securities that aren't needed
// 	// only see different trades -> simulate market that has traders in doing different trades and what is the aggregate position look like
// }
//
// // // DataFeed ... TODO
// // func DataFeed(filename string) error {
// // 	done := make(chan bool)
// //
// // 	cfg, cfgErr := LoadConfig(filename)
// // 	if cfgErr != nil {
// // 		return cfgErr
// // 	}
// // 	// TODO: configurePortfolio(cfg)
// //
// // 	// REVIEW: sort pattern of filepath.Glob?
// // 	fileGlob, globErr := filepath.Glob(cfg.File.Glob)
// // 	if globErr != nil {
// // 		return globErr
// // 	}
// //
// // 	errChan := make(chan error, 1)
// //
// // 	for _, datafile := range fileGlob {
// // 		process(datafile, cfg)
// //
// // 		// tickChan, errChan := loadTicks(recordChan, cfg, durationOp)
// // 	}
// // 	for {
// // 		select {
// // 		case <-done:
// // 			return
// // 		}
// // 	}
// // 	return nil
// // }
//
// func process(done chan<- bool, file string, cfg *Config) error {
// 	quit := make(chan bool)
// 	recordChan := make()
// 	tickChan := make(chan *Tick)
//
// 	datafile, fileErr := os.Open(file)
// 	defer datafile.Close()
// 	if fileErr != nil {
// 		log.Fatal("File cannot be loaded")
// 		return fileErr
// 	}
//
// 	durationOp, durErr := func(filename string) (*simDuration, error) {
// 		fileNoExt := strings.TrimSuffix(filename, filepath.Ext(filename))
//
// 		lastDate, dateErr := time.Parse(cfg.File.ExampleDate, fileNoExt)
// 		if dateErr != nil {
// 			return nil, dateErr
// 		}
// 		return &simDuration{
// 			// HACK: manually setting date
// 			lastDate: lastDate,
// 			barRate:  cfg.Simulation.BarRate,
// 		}, nil
// 	}(file)
// 	if durErr != nil {
// 		return durErr
// 	}
//
// 	r := csv.NewReader(datafile)
// 	r.Comma = cfg.File.Delim
//
// 	go loadRecords(recordChan, quit, r)
//
// 	for {
// 		select {
// 		case record := <-recordChan:
// 			go loadTick(quit, tickChan, record, cfg, durationOp)
// 		case <-quit:
// 			done <- true
// 			return
// 		}
// 	}
//
// }
//
// // loadRecords is a function that generates a channel for reading records from a file.
// func loadRecords(recordChan chan<- []string, quit chan<- bool, r *csv.Reader) {
// 	go func() {
// 		for {
// 			record, err := r.Read()
// 			if err == io.EOF {
// 				quit <- true
// 			}
// 			go loadTick(tickChan, quit, record, cfg, durationOp)
// 		}
// 	}()
// }
//
// // REVIEW: https://play.golang.org/p/Off9_cCiRJ3
// // FIXME: add error channel
// // func loadTick(quit chan<- bool, record []string, cfg *Config, timeOp *simDuration) {
// // 		var tick *Tick
// // 		tick.Ticker = record[cfg.File.Columns.Ticker]
// //
// // 		bid, bidErr := strconv.ParseFloat(record[cfg.File.Columns.Bid], 64)
// // 		if bidErr != nil {
// // 			errChan <- errors.New("Bid Price could not be parsed")
// // 			return nil
// // 		}
// // 		tick.Bid = FloatAmount(bid)
// //
// // 		bidSize, bidSzErr := strconv.ParseFloat(record[cfg.File.Columns.BidSize], 64)
// // 		if bidSzErr != nil {
// // 			errChan <- errors.New("Bid Size could not be parsed")
// // 			return
// // 		}
// // 		tick.BidSize = FloatAmount(bidSize)
// //
// // 		ask, askErr := strconv.ParseFloat(record[cfg.File.Columns.AskSize], 64)
// // 		if askErr != nil {
// // 			errChan <- errors.New("Ask Price could not be parsed")
// // 			return
// // 		}
// // 		tick.Ask = FloatAmount(ask)
// //
// // 		askSize, askSzErr := strconv.ParseFloat(record[cfg.File.Columns.Ask], 64)
// // 		if askSzErr != nil {
// // 			errChan <- errors.New("Ask Size could not be parsed")
// // 			return
// // 		}
// // 		tick.AskSize = FloatAmount(askSize)
// //
// // 		tickDuration, timeErr := time.ParseDuration(record[cfg.File.Columns.Timestamp] + timeOp.timeUnit)
// // 		if timeErr != nil {
// // 			errChan <- timeErr
// // 			return
// // 		}
// // 		tick.Timestamp = timeOp.lastDate.Add(tickDuration)
// // 		out <- tick
// // 	}
// // 	return out, errChan
// // }
//
// func newSimDuraiton(lastDate time.Time, barRate time.Duration) *simDuration {
// 	simDuration := &simDuration{lastDate: lastDate, barRate: barRate}
// 	switch barRate {
// 	case time.Nanosecond:
// 		simDuration.timeUnit = "ns"
// 	case time.Microsecond:
// 		simDuration.timeUnit = "us"
// 	case time.Millisecond:
// 		simDuration.timeUnit = "ms"
// 	case time.Second:
// 		simDuration.timeUnit = "s"
// 	case time.Minute:
// 		simDuration.timeUnit = "m"
// 	case time.Hour:
// 		simDuration.timeUnit = "hr"
// 	}
// 	return simDuration
// }
//
// // TODO processTicks ...
// func processTicks(in <-chan *Tick) {
// 	// go func() {
// 	// 	for tick := range in {
// 	//
// 	// 	}
// 	// }()
//
// }
//
// // func loadTicks(in <-chan []string) <-chan *Tick {
// // 	out := make(chan *Tick)
// // 	var ticksLooped int
// //
// // 	go func() {
// // 		for record := range in {
// // 			tick, ok := createTick(record)
// // 			if ok == true {
// // 				out <- tick
// // 			} else {
// // 				log.Printf("Tick #%d not created due to bad format", ticksLooped)
// // 			}
// // 		}
// // 		close(out)
// // 	}()
// // 	return out
// // }
