package holding

import (
	"fmt"

	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/utils"
)

// Holding structs refer the holding of a financial asset.
type Holding struct {
	Ticker              string
	Volume              utils.Amount
	NumTicks            uint
	AvgBid, AvgAsk      utils.Amount
	LastBid, LastAsk    *utils.DatedMetric
	MaxBid, MaxAsk      *utils.DatedMetric
	MinBid, MinAsk      *utils.DatedMetric
	BuyPrice, SellPrice *utils.DatedMetric
}

func (holding *Holding) String() string {
	return fmt.Sprintf("\nTicker: %s\nVolume: %d\nBuy Price: %d\nDate: %s",
		holding.Ticker, holding.Volume/100, holding.LastBid.Amount, holding.BuyPrice.Date.String())
}

// UpdateMetrics uses tick data to bring its metrics up to date.
func (holding *Holding) UpdateMetrics(tick instrument.Tick) {
	holding.AvgBid = utils.Avg(holding.AvgBid, holding.NumTicks, tick.Bid)
	holding.AvgAsk = utils.Avg(holding.AvgAsk, holding.NumTicks, tick.Ask)
	holding.MaxBid = utils.Max(holding.MaxBid, tick.Bid, tick.Timestamp)
	holding.MinBid = utils.Min(holding.MinBid, tick.Bid, tick.Timestamp)
	holding.MaxAsk = utils.Max(holding.MaxAsk, tick.Ask, tick.Timestamp)
	holding.MinAsk = utils.Min(holding.MinAsk, tick.Ask, tick.Timestamp)
	holding.LastAsk = &utils.DatedMetric{Amount: tick.Ask, Date: tick.Timestamp}
	holding.LastBid = &utils.DatedMetric{Amount: tick.Bid, Date: tick.Timestamp}
	holding.NumTicks++
}
