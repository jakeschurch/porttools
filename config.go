package porttools

import (
	"encoding/json" // will be used later
	"os"
	"time"
)

// NOTE: In a contemporary electronic market (circa 2009), low latency trade processing time was qualified as under 10 milliseconds, and ultra-low latency as under 1 millisecond

// LoadConfig uses a Json File to populate details regarding configuration.
func loadConfig(filename string) (config *Config, err error) {
	file, fileErr := os.Open(filename)
	defer file.Close()
	if fileErr != nil {
		return nil, fileErr
	}
	decoder := json.NewDecoder(file)
	if decodeErr := decoder.Decode(&config); decodeErr != nil {
		return nil, decodeErr
	}
	return config, nil
}

// Config is used as a struct store store configuration data in.
type Config struct {
	File struct {
		Glob        string `json:"fileGlob"`
		Delim       rune   `json:"delim"`
		ExampleDate string `json:"exampleDate"`
		Columns     struct {
			Ticker    int `json:"ticker"`
			Timestamp int `json:"timestamp"`
			Volume    int `json:"volume"`
			Bid       int `json:"bidPrice"`
			BidSize   int `json:"bidSize"`
			Ask       int `json:"askPrice"`
			AskSize   int `json:"askSize"`
		} `json:"columns"`
	} `json:"file"`

	Backtest struct {
		StartCashAmt     float64  `json:"startCashAmt"`
		IgnoreSecurities []string `json:"ignoreSecurities"`
		Slippage         float64  `json:"slippage"`
		Commission       float64  `json:"commission"`
	} `json:"backtest"`

	Simulation struct {
		StartDate  time.Time     `json:"startDate"`
		EndDate    time.Time     `json:"endDate"`
		BarRate    time.Duration `json:"barRate"`
		Costmethod CostMethod    `json:"costmethod"`
		//  IngestRate measures how many bars to skip
		// IngestRate BarDuration `json:"ingestRate"`
	} `json:"simulation"`

	Benchmark struct {
		Use    bool `json:"use"`
		Update bool `json:"update"`
	} `json:"benchmark"`
}

func (cfg *Config) timeUnit() (timeunit string) {
	switch cfg.Simulation.BarRate {
	case time.Nanosecond:
		timeunit = "ns"
	case time.Microsecond:
		timeunit = "us"
	case time.Millisecond:
		timeunit = "ms"
	case time.Second:
		timeunit = "s"
	case time.Minute:
		timeunit = "m"
	case time.Hour:
		timeunit = "hr"
	}
	return
}

// BarDuration is used to register tick intake.
// REVIEW: needed?
type BarDuration time.Duration
