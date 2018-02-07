package porttools

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)
// TODO(DataFeed)
func DataFeed(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("File cannot load")
	}
	r := csv.NewReader(file)
	defer file.Close()

	for {
		record, err := r.Read()

		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return nil
		}
		tick := newTickIn(record)
	}
}

func newTickIn(record []string) *Tick {
	price, _ := strconv.ParseFloat(record[0], 64)
	volume, _ := strconv.ParseFloat(record[1], 64)
	bidSize, _ := strconv.ParseFloat(record[2], 64)
	askSize, _ := strconv.ParseFloat(record[3], 64)
	datetime, _ := time.Parse("Jan 2, 2006 10:04:05.000000000", record[4])

	tick := Tick{
		Ticker: record[5], Price: price, Volume: volume,
		BidSize: bidSize, AskSize: askSize, Datetime: datetime,
	}
	return &tick
}
