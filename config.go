package porttools

import (
	"encoding/json" // will be used later
	"log"
	"os"
	"time"
)

// NOTE: In a contemporary electronic market (circa 2009), low latency trade processing time was qualified as under 10 milliseconds, and ultra-low latency as under 1 millisecond

// LoadConfig uses a Json File to populate details regarding configuration.
func loadConfig(filename string) (*Config, error) {
	var config *Config

	file, fileErr := os.Open(filename)
	defer file.Close()
	if fileErr != nil {
		return nil, fileErr
	}
	decoder := json.NewDecoder(file)
	if decodeErr := decoder.Decode(&config); decodeErr != nil {
		log.Fatal("Could not read config file")
		return nil, decodeErr
	}
	log.Println(config.File.Glob)
	return config, nil
}

// Config is used as a struct store store configuration data in.
type Config struct {
	File struct {
		Glob          string `json:"glob"`
		Headers       bool   `json:"headers"`
		Delim         string `json:"delim"`
		ExampleDate   string `json:"exampleDate"`
		TimestampUnit string `json:"timestampUnit"`

		Columns struct {
			Ticker    int `json:"ticker"`
			Timestamp int `json:"timestamp"`
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
		StartDate  string        `json:"startDate"`
		EndDate    string        `json:"endDate"`
		BarRate    time.Duration `json:"barRate"`
		Costmethod CostMethod    `json:"costmethod"`
		// TODO: REVIEW good idea to use go generate for output format and other consts?
		OutFmt OutputFmt `json:"outFmt"`
		//  IngestRate measures how many bars to skip
		// IngestRate BarDuration `json:"ingestRate"`
	} `json:"simulation"`

	Benchmark struct {
		Use    bool `json:"use"`
		Update bool `json:"update"`
	} `json:"benchmark"`
}

// BarDuration is used to register tick intake.
// REVIEW: needed?
type BarDuration time.Duration
