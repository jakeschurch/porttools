package instrument

import (
	"time"

	"github.com/jakeschurch/porttools/utils"
	"github.com/pkg/errors"
)

// ------------------------------------------------------------------
var (
	ErrNegativeVolume = errors.New("cannot have negative volume")

	// ErrZeroVolume indicates a struct has a volume of 0 and should be deleted entirely.
	ErrZeroVolume = errors.New("struct has volume of 0, struct no longer active")
)

// ------------------------------------------------------------------

// Instrument is the base type of a financial widget.
type Instrument interface {
	Update(data Quote)
}

func ExtractOrder(i Instrument) (order *Order) {
	switch x := i.(type) {
	case *Order:
		return x
	case Order:
		return &x
	}
	return nil
}

func ExtractQuote(i Instrument) (quote *Quote) {
	switch x := i.(type) {
	case Quote:
		quote = &x

	case *Quote:
		quote = x

	case *Order:
		quote = x.Quote

	case *Holding:
		quote = x.Quote

	case *Security:
		quote = x.Quote
	}
	return quote
}

// SellOff a struct that satisfies the Instrument interface.
// Returns a new Security struct and a pointer to the updated under send signal to remove if volume of 0.
func SellOff(i Instrument, o Order, summ *AssetSumm) (*Security, error) {
	var v = o.Volume
	var sellData = &utils.DatedMetric{Amount: o.Ask, Date: o.Timestamp}

	var quote, ok = i.(*Quote)
	if !ok {
		return nil, errors.New("quote cannot satisfy Instrument interface")
	}

	if quote.Volume -= o.Volume; quote.Volume == 0 {
		return nil, errors.Wrap(ErrNegativeVolume, "of quote")
	}
	if summ.TotalVolume -= v; summ.TotalVolume == 0 {
		return nil, errors.Wrap(ErrNegativeVolume, "of assetSumm")
	}

	return NewSecurity(sellData, NewHolding(*quote), summ), nil
}

// ------------------------------------------------------------------

type Quote struct {
	Ticker           string
	Volume, Bid, Ask utils.Amount
	Timestamp        time.Time
	Nseen            uint // used when
}

func NewQuote(ticker string, volume, bid, ask utils.Amount, ts time.Time) *Quote {
	return &Quote{
		Ticker: ticker, Volume: volume,
		Bid: bid, Ask: ask, Timestamp: ts,
		Nseen: 1,
	}
}

// Update quote's ask/bid fields based on other quote.
func (q Quote) Update(data Quote) {
	q.Ask = data.Ask
	q.Bid = data.Bid
}

// ------------------------------------------------------------------

// Tick structs holds information about a financial asset at a specific point in time.
type Tick struct {
	*Quote
	BidSize, AskSize utils.Amount
}

// NewTick creates new Tick.
func NewTick(bidSz, askSz utils.Amount, q *Quote) *Tick {
	return &Tick{
		Quote:   q,
		BidSize: bidSz, AskSize: askSz,
	}
}

// Update Tick from quote data.
func (t Tick) Update(data Quote) {
	t.Quote.Update(data)
}

// ------------------------------------------------------------------

// AssetSumm reflects an Insturment's trading metrics.
type AssetSumm struct {
	Nticks                      uint
	TotalVolume, AvgBid, AvgAsk utils.Amount

	MaxBid, MaxAsk *utils.DatedMetric
	MinBid, MinAsk *utils.DatedMetric
}

// NewAssetSumm instantiaties a new struct of type AssetSumm.
func NewAssetSumm(q Quote) *AssetSumm {
	assetBid := &utils.DatedMetric{Amount: q.Bid, Date: q.Timestamp}
	assetAsk := &utils.DatedMetric{Amount: q.Ask, Date: q.Timestamp}

	return &AssetSumm{
		Nticks: 1, TotalVolume: q.Volume,
		AvgBid: assetBid.Amount, AvgAsk: assetAsk.Amount,
		MaxBid: assetBid, MaxAsk: assetAsk,
		MinBid: assetBid, MinAsk: assetAsk,
	}
}

// Update Asset Summary metrics from quote data.
func (a AssetSumm) Update(q Quote) {
	// update bid metrics
	a.AvgBid = utils.Avg(a.AvgBid, a.Nticks, q.Bid)
	a.MaxBid = utils.Max(a.MaxBid, q.Bid, q.Timestamp)
	a.MinBid = utils.Min(a.MinBid, q.Bid, q.Timestamp)

	// update ask metrics
	a.AvgAsk = utils.Avg(a.AvgAsk, a.Nticks, q.Ask)
	a.MaxAsk = utils.Max(a.MaxAsk, q.Ask, q.Timestamp)
	a.MinAsk = utils.Min(a.MinAsk, q.Ask, q.Timestamp)

	a.Nticks++
}

// GetVolume returns total Volume.
func (a AssetSumm) GetVolume() utils.Amount {
	return a.TotalVolume
}

// ------------------------------------------------------------------

// Holding structs refer the holding of a financial asset.
type Holding struct {
	*Quote
}

// NewHolding instantities struct of type Holding.
// Map Buy -> Bid, Sell -> Ask
// Only open holdings/positions are allowed to be this type.
// When partially/all sold off, becomes a struct.
func NewHolding(q Quote) *Holding {
	return &Holding{Quote: &q}
}

// ------------------------------------------------------------------

// Security structs hold information regarding a financial asset for the entire
// life of the financial asset in a trading environment. Because a Security struct
// holds aggregate information regarding a financial asset, it is embedded into an Index or Benchmark.
type Security struct {
	*Holding
	*AssetSumm
	SellDate time.Time
}

// NewSecurity instantiates a security object from Tick data.
func NewSecurity(sellAt *utils.DatedMetric, h *Holding, summ *AssetSumm) *Security {
	s := &Security{
		Holding:   h,
		AssetSumm: summ,
		SellDate:  sellAt.Date,
	}
	s.Ask = sellAt.Amount
	s.Nticks = s.Nticks - h.Nseen

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

func (o Order) Update(data Quote) {
	o.Quote.Update(data)
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
