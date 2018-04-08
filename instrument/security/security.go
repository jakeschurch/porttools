package security

import (
	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/utils"
)

// Security structs hold information regarding a financial asset for the entire
// life of the financial asset in a trading environment. Because a Security struct
// holds aggregate information regarding a financial asset, it is embedded into an Index or Benchmark.
type Security struct {
	Ticker              string
	NumTicks            uint
	BuyPrice, SellPrice *utils.DatedMetric

	// Fields related to bid/ask prices
	AvgBid, AvgAsk   utils.Amount
	LastBid, LastAsk *utils.DatedMetric
	MaxAsk, MinAsk   *utils.DatedMetric
	MaxBid, MinBid   *utils.DatedMetric
	// Fields related to bid/ask volume
	AvgBidSize, AvgAskSize utils.Amount
	MaxAskSize, MinAskSize *utils.DatedMetric
	MaxBidSize, MinBidSize *utils.DatedMetric
}

// New instantiates a security object from Tick data.
func New(tick instrument.Tick) *Security {
	firstBid := &utils.DatedMetric{Amount: tick.Bid, Date: tick.Timestamp}
	firstAsk := &utils.DatedMetric{Amount: tick.Ask, Date: tick.Timestamp}
	firstBidSize := &utils.DatedMetric{Amount: tick.BidSize, Date: tick.Timestamp}
	firstAskSize := &utils.DatedMetric{Amount: tick.AskSize, Date: tick.Timestamp}

	return &Security{
		Ticker: tick.Ticker, NumTicks: 1,
		AvgBid: tick.Bid, AvgAsk: tick.Ask,
		MaxBid: firstBid, MaxAsk: firstAsk,
		MinBid: firstBid, MinAsk: firstAsk,
		BuyPrice:   firstAsk,
		AvgBidSize: tick.BidSize, AvgAskSize: tick.AskSize,
		MaxBidSize: firstBidSize, MaxAskSize: firstAskSize,
		MinBidSize: firstBidSize, MinAskSize: firstAskSize,
	}
}

// UpdateMetrics ..TODO
func (s *Security) UpdateMetrics(tick instrument.Tick) {

	s.AvgBid = utils.Avg(s.AvgBid, s.NumTicks, tick.Bid)
	s.AvgAsk = utils.Avg(s.AvgAsk, s.NumTicks, tick.Ask)
	s.AvgBidSize = utils.Avg(s.AvgBid, s.NumTicks, tick.Bid)
	s.AvgAskSize = utils.Avg(s.AvgAsk, s.NumTicks, tick.Ask)
	s.LastAsk = &utils.DatedMetric{Amount: tick.Ask, Date: tick.Timestamp}
	s.LastBid = &utils.DatedMetric{Amount: tick.Bid, Date: tick.Timestamp}
	s.MaxBid = utils.Max(s.MaxBid, tick.Bid, tick.Timestamp)
	s.MinBid = utils.Min(s.MinBid, tick.Bid, tick.Timestamp)
	s.MaxBidSize = utils.Max(s.MaxBidSize, tick.BidSize, tick.Timestamp)
	s.MinBidSize = utils.Min(s.MinBidSize, tick.BidSize, tick.Timestamp)
	s.MaxAsk = utils.Max(s.MaxAsk, tick.Ask, tick.Timestamp)
	s.MinAsk = utils.Min(s.MinAsk, tick.Ask, tick.Timestamp)
}
