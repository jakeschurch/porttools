package security

import (
	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/utils"
)

// Security structs hold information regarding a financial asset for the entire
// life of the financial asset in a trading environment. Because a Security struct
// holds aggregate information regarding a financial asset, it is embedded into an Index or Benchmark.
type Security struct {
	*instrument.Asset
	BuyPrice, SellPrice *utils.DatedMetric
}

// New instantiates a security object from Tick data.
func New(buy, sell *utils.DatedMetric, asset instrument.Asset) *Security {
	return &Security{
		Asset:    &asset,
		BuyPrice: buy, SellPrice: sell,
	}
}

// GetUnderlying method for security returns an instrument.Asset type.
func (s Security) GetUnderlying() instrument.Financial {
	return s.Asset
}
