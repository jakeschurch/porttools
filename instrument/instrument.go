package instrument

import (
	"time"

	"github.com/jakeschurch/porttools/utils"
)

// Tick structs holds information about a financial asset at a specific point in time.
type Tick struct {
	Ticker           string
	Bid, Ask         utils.Amount
	BidSize, AskSize utils.Amount
	Timestamp        time.Time
}

// Instrument is the base type of a financial widget.
type Instrument struct {
	Ticker string
	Volume utils.Amount
}

// NewInstrument instantities a new struct of type Instrument.
func NewInstrument(ticker string, volume utils.Amount) *Instrument {
	return &Instrument{ticker, volume}
}

// Asset is tradeable instrument type.
type Asset struct {
	*Instrument
	nTicks uint

	AvgBid, AvgAsk   utils.Amount
	LastBid, LastAsk *utils.DatedMetric
	MaxBid, MaxAsk   *utils.DatedMetric
	MinBid, MinAsk   *utils.DatedMetric
}

// NewAsset instantiaties a new struct of type Asset.
func NewAsset(ticker string, bid, ask, volume utils.Amount, timestamp time.Time) *Asset {
	assetBid := &utils.DatedMetric{Amount: bid, Date: timestamp}
	assetAsk := &utils.DatedMetric{Amount: ask, Date: timestamp}

	return &Asset{
		Instrument: &Instrument{Ticker: ticker, Volume: volume},

		nTicks: 1,
		AvgBid: assetBid.Amount, AvgAsk: assetAsk.Amount,
		LastBid: assetBid, MaxBid: assetBid, MinBid: assetBid,
		LastAsk: assetAsk, MaxAsk: assetAsk, MinAsk: assetAsk,
	}
}

// Update uses new tick data to update an asset's metrics.
func (a *Asset) Update(t *Tick) {
	// update bid metrics
	a.AvgBid = utils.Avg(a.AvgBid, a.nTicks, t.Bid)
	a.LastBid = &utils.DatedMetric{Amount: t.Bid, Date: t.Timestamp}
	a.MaxBid = utils.Max(a.MaxBid, t.Bid, t.Timestamp)
	a.MinBid = utils.Min(a.MinBid, t.Bid, t.Timestamp)

	// update ask metrics
	a.AvgAsk = utils.Avg(a.AvgAsk, a.nTicks, t.Ask)
	a.LastAsk = &utils.DatedMetric{Amount: t.Ask, Date: t.Timestamp}
	a.MaxAsk = utils.Max(a.MaxAsk, t.Ask, t.Timestamp)
	a.MinAsk = utils.Min(a.MinAsk, t.Ask, t.Timestamp)

	a.nTicks++
}
