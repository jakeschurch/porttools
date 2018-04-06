package config

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/jakeschurch/porttools/output"
	"github.com/jakeschurch/porttools/utils"
)

// NOTE: In a contemporary electronic market (circa 2009), low latency trade processing time was qualified as under 10 milliseconds, and ultra-low latency as under 1 millisecond

// Load returns a config item.
func Load(filename string) (*Config, error) {
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
			Ticker    uint8 `json:"ticker"`
			Timestamp uint8 `json:"timestamp"`
			Bid       uint8 `json:"bid"`
			BidSize   uint8 `json:"bidSize"`
			Ask       uint8 `json:"ask"`
			AskSize   uint8 `json:"askSize"`
		} `json:"columns"`
	} `json:"file"`

	Backtest struct {
		StartCashAmt     float64  `json:"startCashAmt"`
		IgnoreSecurities []string `json:"ignoreSecurities"`
		Slippage         float64  `json:"slippage"`
		Commission       float64  `json:"commission"`
	} `json:"backtest"`

	Simulation struct {
		StartDate  string           `json:"startDate"`
		EndDate    string           `json:"endDate"`
		BarRate    time.Duration    `json:"barRate"`
		Costmethod utils.CostMethod `json:"costmethod"`
		// TODO: REVIEW good idea to use go generate for output format and other consts?
		OutFmt output.Fmt `json:"outFmt"`
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
