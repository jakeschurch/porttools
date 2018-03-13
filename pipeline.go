package porttools

import (
	"encoding/csv"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func notMain() {

	// example architecture
	// 	stream <- Tick
	// 	map[Tick.Ticker] <- stream
	// 	process <- map
	// ticksIn := make(map[string](chan Tick))
	//
	// doneProcessing := false
	// for doneProcessing == false {
	// 	// Loop over file
	// 	// store each record in a Tick reference and pass it to ticksIn
	// 	exampleTick := Tick{Ticker: "AAPL", bid: 101.14, Volume: 20000.00, Datetime: time.Now()}

	// go func() {
	// 	if recordChannel, exists := ticksIn[exampleTick.Ticker]; !exists {
	// 		ticksIn[exampleTick.Ticker] = make(chan Tick)
	// 	}
	// 	ticksIn[exampleTick.Ticker] <- exampleTick
	// }()
	// option to throw out securities that aren't needed
	// only see different trades -> simulate market that has traders in doing different trades and what is the aggregate position look like
}

// DataFeed ... TODO
func DataFeed(filename string) error {

	cfg, cfgErr := LoadConfig(filename)
	if cfgErr != nil {
		return cfgErr
	}
	// TODO: add pre-process portfolio function

	// REVIEW: sort pattern of filepath.Glob?
	fileGlob, globErr := filepath.Glob(cfg.File.Glob)
	if globErr != nil {
		return globErr
	}

	fileNoExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	durationOp := &simDuration{
		// HACK: manually setting date
		lastDate: time.Parse(cfg.File.ExampleDate, fileNoExt),
		barRate:  cfg.Simulation.BarRate,
	}

	for _, dataFile := range fileGlob {
		recordFile, fileErr := os.Open(dataFile)
		defer recordFile.Close()

		if fileErr != nil {
			log.Fatal("File cannot be loaded")
		}

		r := csv.NewReader(recordFile)
		r.Comma = cfg.File.Delim

		recordChan := loadRecords(r)
		tickChan := loadTicks(recordChan, cfg)

	}

	return nil
}

func newSimDuraiton(lastDate time.Time, barRate BarDuration) *simDuration {
	simDuration := &simDuration{lastDate: lastDate, barRate: barRate}
	switch barRate {
	case time.Nanosecond:
		simDuration.timeUnit = "ns"
	case time.Microsecond:
		simDuration.timeUnit = "us"
	case time.Millisecond:
		simDuration.timeUnit = "ms"
	case time.Second:
		simDuration.timeUnit = "s"
	case time.Minute:
		simDuration.timeUnit = "m"
	case time.Hour:
		simDuration.timeUnit = "hr"
	}
	return simDuration
}

type simDuration struct {
	lastDate time.Time
	barRate  BarDuration
	timeUnit string
}

// func loadRecords(r *csv.Reader) <-chan []string {
// 	errChan := make(chan error, 1)
// 	out := make(chan []string)
//
// 	go func() {
// 		for {
// 			record, err := r.Read()
// 			if err == io.EOF {
// 				break
// 			}
// 			out <- record
// 		}
// 		close(out)
// 	}()
// 	return out
// }

// REVIEW: https://play.golang.org/p/Off9_cCiRJ3

func loadTicks(recordChan <-chan []string, cfg *Config, timeOp *simDuration) chan<- *Tick {
	out := make(chan *Tick)
	go func() {
		for record := range recordChan {
			var tick *Tick
			tick.Ticker = record[cfg.File.Columns.Ticker]

			bidPrice, bidErr := strconv.ParseFloat(record[cfg.File.Columns.BidPrice], 64)
			if bidErr != nil {
				return nil, errors.New("Bid Price could not be parsed")
			}
			tick.BidPrice = FloatAmount(bidSize)

			bidSize, bidSzErr := strconv.ParseFloat(record[cfg.File.Columns.BidSize], 64)
			if bidSzErr != nil {
				return nil, errors.New("Bid Size could not be parsed")
			}
			tick.BidSize = FloatAmount(bidSize)

			askPrice, askErr := strconv.ParseFloat(record[cfg.File.Columns.AskSize], 64)
			if askErr != nil {
				return nil, errors.New("Ask Price could not be parsed")
			}
			tick.askPrice = FloatAmount(askSize)

			askSize, askSzErr := strconv.ParseFloat(record[cfg.File.Columns.AskPrice], 64)
			if askSzErr != nil {
				return nil, errors.New("Ask Size could not be parsed")
			}
			tick.askSize = FloatAmount(askSize)

			tickDuration, timeErr := time.ParseDuration(record[cfg.File.Columns.Timestamp] + timeOp.timeUnit)
			if timeErr != nil {
				return nil, !ok
			}
			tick.Timestamp = timeOp.lastDate.Add(tickDuration)
			out <- tick
		}
		close(out)
	}()
	return out
}

// TODO processTicks ...
func processTicks(in <-chan *Tick) {
	// go func() {
	// 	for tick := range in {
	//
	// 	}
	// }()

}

// func loadTicks(in <-chan []string) <-chan *Tick {
// 	out := make(chan *Tick)
// 	var ticksLooped int
//
// 	go func() {
// 		for record := range in {
// 			tick, ok := createTick(record)
// 			if ok == true {
// 				out <- tick
// 			} else {
// 				log.Printf("Tick #%d not created due to bad format", ticksLooped)
// 			}
// 		}
// 		close(out)
// 	}()
// 	return out
// }
