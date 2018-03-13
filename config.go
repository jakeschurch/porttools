package porttools

import (
	_ "encoding/json"
	_ "flag"
	"path/filepath"
	"time"
)

// TODO: LoadConfig ...
func LoadConfig(filename string) interface{} {
	return nil
}

type benchmarkConfig struct {
	Load         bool     `json:"load"`
	Update       bool     `json:"update"`
	Constituents []string `json:"constituents"`
}

type backtestConfig struct {
	slippagePerTrade Amount
	commission       Amount
}

type simConfig struct {
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	BarRate   Bar       `json:"barRate"`
	//  IngestRate measures how many bars to skip
	IngestRate Bar `json:"ingestRate"`
}

// Bar type is used to register tick intake.
type Bar uint

const (
	microsecond Bar = iota
	second
	minute
	hour
	day
)

// QUESTION: is this function needed?
func (cfg simConfig) dataFiles(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
	// QUESTION: is this if statement necessary if Glob is creating
	// 		error for us?
	// if files, err := filepath.Glob(pattern); err != nil {
	// 	return files, err
	// } else {
	// 	return files, nil
	// }
}

type dataConfig struct {
	FileGlob   string
	tickerCol  uint
	volumeCol  uint
	bidAmtCol  uint
	bidSizeCol uint
	askAmtCol  uint
	askSizeCol uint
}
