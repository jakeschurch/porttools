package holding

import (
	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/utils"
)

// Holding structs refer the holding of a financial asset.
type Holding struct {
	*instrument.Instrument
	BuyPrice, SellPrice *utils.DatedMetric
}

// New instantities struct of type Holding.
func New(instrument *instrument.Instrument, buyPrice *utils.DatedMetric) *Holding {
	return &Holding{
		Instrument: instrument,
		BuyPrice:   buyPrice,
	}
}

// func (holding *Holding) String() string {
// 	return fmt.Sprintf("\nTicker: %s\nVolume: %d\nBuy Price: %d\nDate: %s",
// 		holding.Ticker, holding.Volume/100, holding.LastBid.Amount, holding.BuyPrice.Date.String())
// }
