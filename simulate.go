package porttools

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	file, err := os.Open("test.py")
	if err != nil {
		log.Fatal("File cannot load")
	}
	r := csv.NewReader(file)
	defer file.Close()

	record, err := r.Read()
	for err != io.EOF {
		go func() { fmt.Println(record) }()
	}

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
	// 	exampleTick := Tick{Ticker: "AAPL", Price: 101.14, Volume: 20000.00, Datetime: time.Now()}

	// go func() {
	// 	if inChannel, exists := ticksIn[exampleTick.Ticker]; !exists {
	// 		ticksIn[exampleTick.Ticker] = make(chan Tick)
	// 	}
	// 	ticksIn[exampleTick.Ticker] <- exampleTick
	// }()
	// option to throw out securities that arent needed
	// only see different trades -> simulate market that has traders in doing different trades and what is the aggregate position look like
}

// DataFeed TODO
func DataFeed(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("File cannot load")
	}
	r := csv.NewReader(file)
	defer file.Close()

	inChan := make(chan []string)
	tickChan := make(chan *Tick)

	go func() {
		for {
			if record, err := r.Read(); err != nil {
				if err == io.EOF {
					break
				}
				inChan <- record
			}

			go func(inChan <-chan *Tick) {
				for range inChan {
					go processTick(<-inChan)
				}
			}(tickChan)
		}
	}()

	return nil
}

func createTick(record []string, outChan chan<- *Tick) error {
	tick := new(Tick)

	if price, err := strconv.ParseFloat(record[0], 64); err != nil {
		return err
	} else {
		tick.Price = price
	}

	if volume, err := strconv.ParseFloat(record[1], 64); err != nil {
		return err
	} else {
		tick.Volume = volume
	}

	if bidSize, err := strconv.ParseFloat(record[2], 64); err != nil {
		return err
	} else {
		tick.BidSize = bidSize
	}

	if askSize, err := strconv.ParseFloat(record[3], 64); err != nil {
		return err
	} else {
		tick.AskSize = askSize
	}

	if datetime, err := time.Parse("Jan 2, 2006 10:04:05.000000000", record[4]); err != nil {
		return err
	} else {
		tick.Datetime = datetime
	}

	outChan <- tick
	return nil
}

func processTick(tick *Tick) (ok bool) {
	// for tick := range tickChan {

	//
	// }
	return true
}
