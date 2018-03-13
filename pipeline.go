package porttools

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
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

// DataFeed TODO
func DataFeed(fileName string) error {
	// recordCh := loadRecords(fileName)
	// loadTicks(tickCh, recordCh)
	// processTicks(tickCh)
	//
	// go func() {
	// 	for {
	// 		select {
	// 		case msg := <-recordCh:
	// 			loadTicks(msg)
	// 		}
	// 	}
	// }()

	// 	// implement goroutine for []string
	// 	go func(recordChan <-chan *Tick) {
	// 		for range recordChan {
	// 			go processTick(<-recordChan)
	// 		}
	// 	}(tickChan)
	// }

	return nil
}

func loadRecords(fileName string) <-chan []string {
	out := make(chan []string)

	file, err := os.Open(fileName)
	defer file.Close()

	if err != nil {
		log.Fatal("File cannot load")
	}
	r := csv.NewReader(file)

	go func() {
		for {
			if record, err := r.Read(); err != nil {
				if err == io.EOF {
					break
				}
			} else {
				out <- record
			}
		}
		close(out)
	}()
	return out
}

func loadTicks(in <-chan []string) <-chan *Tick {
	out := make(chan *Tick)
	var ticksLooped int

	go func() {
		for record := range in {
			tick, ok := createTick(record)
			if ok == true {
				out <- tick
			} else {
				log.Printf("Tick #%d not created due to bad format", ticksLooped)
			}
		}
		close(out)
	}()
	return out
}

func createTick(record []string) (tick *Tick, ok bool) {
	ok = true

	if bid, bidErr := strconv.ParseUint("test", 10, 64); bidErr == nil {
		tick.BidSize = Amount(bid)
	} else {
		return nil, !ok
	}
	if volume, VolumeErr := strconv.ParseUint("test", 10, 64); VolumeErr == nil {
		tick.Volume = Amount(volume)
	} else {
		return nil, !ok
	}
	if bid, bidErr := strconv.ParseUint("test", 10, 64); bidErr == nil {
		tick.BidSize = Amount(bid)
	} else {
		return nil, !ok
	}
	if ask, askErr := strconv.ParseUint("test", 10, 64); askErr == nil {
		tick.AskSize = Amount(ask)
	} else {
		return nil, !ok
	}
	if datetime, dateErr := time.Parse("Jan 2, 2006 100405.000000000", record[4]); dateErr != nil {
		tick.Datetime = datetime
	} else {
		return nil, !ok
	}
	return tick, ok
}

// TODO processTicks ...
func processTicks(in <-chan *Tick) {
	// go func() {
	// 	for tick := range in {
	//
	// 	}
	// }()

}
