package instrument

import (
	"time"

	"github.com/jakeschurch/porttools/utils"
)

// ------------------------------------------------------------------

// Instrument is the base type of a financial widget.
type Instrument interface {
	Update(Quote)
}

// ------------------------------------------------------------------

type Quote struct {
	Ticker           string
	Volume, Bid, Ask utils.Amount
	Timestamp        time.Time
}

func NewQuote(ticker string, volume, bid, ask utils.Amount, ts time.Time) *Quote {
	return &Quote{
		Ticker: ticker, Volume: volume,
		Bid: bid, Ask: ask, Timestamp: ts,
	}
}

func (q *Quote) Update(fed *Quote) {
	q.Ask = fed.Ask
	q.Bid = fed.Bid
}

// ------------------------------------------------------------------

// Tick structs holds information about a financial asset at a specific point in time.
type Tick struct {
	*Quote
	BidSize, AskSize utils.Amount
}

func NewTick(bidSz, askSz utils.Amount, q *Quote) *Tick {
	return &Tick{
		Quote:   q,
		BidSize: bidSz, AskSize: askSz,
	}
}

func (t *Tick) Update(fed *Quote) {
	t.Update(fed)
}

// ------------------------------------------------------------------

// AssetSumm reflects an Insturment's trading metrics.
type AssetSumm struct {
	nTicks                      uint
	TotalVolume, AvgBid, AvgAsk utils.Amount

	MaxBid, MaxAsk *utils.DatedMetric
	MinBid, MinAsk *utils.DatedMetric
}

// NewAssetSumm instantiaties a new struct of type AssetSumm.
func NewAssetSumm(q *Quote) *AssetSumm {
	assetBid := &utils.DatedMetric{Amount: q.Bid, Date: q.Timestamp}
	assetAsk := &utils.DatedMetric{Amount: q.Ask, Date: q.Timestamp}

	return &AssetSumm{
		nTicks: 1, TotalVolume: q.Volume,
		AvgBid: assetBid.Amount, AvgAsk: assetAsk.Amount,
		MaxBid: assetBid, MaxAsk: assetAsk,
		MinBid: assetBid, MinAsk: assetAsk,
	}
}

func (a *AssetSumm) Update(q Quote) {
	// update bid metrics
	a.AvgBid = utils.Avg(a.AvgBid, a.nTicks, q.Bid)
	a.MaxBid = utils.Max(a.MaxBid, q.Bid, q.Timestamp)
	a.MinBid = utils.Min(a.MinBid, q.Bid, q.Timestamp)

	// update ask metrics
	a.AvgAsk = utils.Avg(a.AvgAsk, a.nTicks, q.Ask)
	a.MaxAsk = utils.Max(a.MaxAsk, q.Ask, q.Timestamp)
	a.MinAsk = utils.Min(a.MinAsk, q.Ask, q.Timestamp)

	a.nTicks++
}

// ------------------------------------------------------------------

// Holding structs refer the holding of a financial asset.
type Holding struct {
	*Quote
	nSeen uint
}

// NewHolding instantities struct of type Holding.
// Map Buy -> Bid, Sell -> Ask
// Only open holdings/positions are allowed to be this type.
// When partially/all sold off, becomes a struct.
func NewHolding(q *Quote) *Holding {
	return &Holding{
		Quote: q,
		nSeen: 0,
	}
}

// SellOff a holding, return a new Security; send signal to remove if volume of 0.
func (h *Holding) SellOff(v utils.Amount, summ *AssetSumm, sellData *utils.DatedMetric) (*Security, bool) {
	var signal bool
	var security = NewSecurity(sellData, h, summ)

	summ.TotalVolume -= v
	h.Volume -= v
	if h.Volume == 0 {
		signal = true
	}
	signal = false

	return security, signal
}

// ------------------------------------------------------------------

// Security structs hold information regarding a financial asset for the entire
// life of the financial asset in a trading environment. Because a Security struct
// holds aggregate information regarding a financial asset, it is embedded into an Index or Benchmark.
type Security struct {
	Holding
	AssetSumm
	SellDate time.Time
}

// NewSecurity instantiates a security object from Tick data.
func NewSecurity(sellAt *utils.DatedMetric, h *Holding, summ *AssetSumm) *Security {
	s := &Security{
		Holding:   *h,
		AssetSumm: *summ,
		SellDate:  sellAt.Date,
	}
	s.Ask = sellAt.Amount
	s.nTicks = s.nTicks - h.nSeen

	return s
}

// ------------------------------------------------------------------

// Order type represents the transaction of a holding
type Order struct {
	*Quote
	Buy    bool // it's either a buy or sell
	Status Status
	Logic  Logic
}

// NewOrder returns a new order that will execute at nearest price.
func NewOrder(buy bool, q *Quote) *Order {
	return &Order{
		Quote:  q,
		Buy:    buy,
		Status: open,
		Logic:  market, // TEMP: for now, only accepting market orders
	}
}

// Status variables refer to a status of an order's execution.
type Status int

const (
	open Status = iota // 0
	closed
	cancelled
	expired // 3
)

// Logic is used to identify when the order should be executed.
type Logic int

const (
	market Logic = iota // 0
	limit
	stopLimit
	stopLoss
	dayTrade // 4
)
