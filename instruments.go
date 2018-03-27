// Package porttools allows for storage of information regarding particular securities.
package porttools

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// Security structs hold information regarding a financial asset for the entire
// life of the financial asset in a trading environment. Because a Security struct
// holds aggregate information regarding a financial asset, it is embedded into an Index or Benchmark.
type Security struct {
	Ticker              string
	NumTicks            uint
	BuyPrice, SellPrice *datedMetric

	// Fields related to bid/ask prices
	AvgBid, AvgAsk   Amount
	LastBid, LastAsk *datedMetric
	MaxAsk, MinAsk   *datedMetric
	MaxBid, MinBid   *datedMetric
	// Fields related to bid/ask volume
	AvgBidSize, AvgAskSize Amount
	MaxAskSize, MinAskSize *datedMetric
	MaxBidSize, MinBidSize *datedMetric
}

// NewSecurity instantiates a new security from Tick data.
func NewSecurity(tick Tick) *Security {
	firstBid := &datedMetric{Amount: tick.Bid, Date: tick.Timestamp}
	firstAsk := &datedMetric{Amount: tick.Ask, Date: tick.Timestamp}
	firstBidSize := &datedMetric{Amount: tick.BidSize, Date: tick.Timestamp}
	firstAskSize := &datedMetric{Amount: tick.AskSize, Date: tick.Timestamp}

	return &Security{
		Ticker: tick.Ticker, NumTicks: 1,
		AvgBid: tick.Bid, AvgAsk: tick.Ask,
		MaxBid: firstBid, MaxAsk: firstAsk,
		MinBid: firstBid, MinAsk: firstAsk,
		AvgBidSize: tick.BidSize, AvgAskSize: tick.AskSize,
		MaxBidSize: firstBidSize, MaxAskSize: firstAskSize,
		MinBidSize: firstBidSize, MinAskSize: firstAskSize,
	}
}

func (s *Security) updateMetrics(tick Tick) {
	s.AvgBid = newAvg(s.AvgBid, s.NumTicks, tick.Bid)
	s.AvgAsk = newAvg(s.AvgAsk, s.NumTicks, tick.Ask)
	s.AvgBidSize = newAvg(s.AvgBid, s.NumTicks, tick.Bid)
	s.AvgAskSize = newAvg(s.AvgAsk, s.NumTicks, tick.Ask)
	s.LastAsk = &datedMetric{tick.Ask, tick.Timestamp}
	s.LastBid = &datedMetric{tick.Bid, tick.Timestamp}
	s.MaxBid = newMax(s.MaxBid, tick.Bid, tick.Timestamp)
	s.MinBid = newMin(s.MinBid, tick.Bid, tick.Timestamp)
	s.MaxBidSize = newMax(s.MaxBidSize, tick.BidSize, tick.Timestamp)
	s.MinBidSize = newMin(s.MinBidSize, tick.BidSize, tick.Timestamp)
	s.MaxAsk = newMax(s.MaxAsk, tick.Ask, tick.Timestamp)
	s.MinAsk = newMin(s.MinAsk, tick.Ask, tick.Timestamp)
}

// Position structs refer the holding of a financial asset.
type Position struct {
	Ticker              string
	Volume              Amount
	NumTicks            uint
	AvgBid, AvgAsk      Amount
	LastBid, LastAsk    *datedMetric
	MaxBid, MaxAsk      *datedMetric
	MinBid, MinAsk      *datedMetric
	BuyPrice, SellPrice *datedMetric
}

func (pos *Position) String() string {
	return fmt.Sprintf("\nTicker: %s\nVolume: %d\nBuy Price: %d\nDate: %s",
		pos.Ticker, pos.Volume/100, pos.LastBid.Amount, pos.BuyPrice.Date.String())
}

func (pos *Position) updateMetrics(tick *Tick) {
	pos.AvgBid = newAvg(pos.AvgBid, pos.NumTicks, tick.Bid)
	pos.AvgAsk = newAvg(pos.AvgAsk, pos.NumTicks, tick.Ask)
	pos.MaxBid = newMax(pos.MaxBid, tick.Bid, tick.Timestamp)
	pos.MinBid = newMin(pos.MinBid, tick.Bid, tick.Timestamp)
	pos.MaxAsk = newMax(pos.MaxAsk, tick.Ask, tick.Timestamp)
	pos.MinAsk = newMin(pos.MinAsk, tick.Ask, tick.Timestamp)
	pos.LastAsk = &datedMetric{tick.Ask, tick.Timestamp}
	pos.LastBid = &datedMetric{tick.Bid, tick.Timestamp}
	pos.NumTicks++
}

// Tick structs holds information about a financial asset at a specific point in time.
type Tick struct {
	Ticker           string
	Bid, Ask         Amount
	BidSize, AskSize Amount
	Timestamp        time.Time
}

// Amount is a representation of fractional volumes. To get around floating-point erroneous behavior, multiply volume by 100 and cap it as an integer.
type Amount int64

// FloatAmount converts a float64 value to an amount type. Can be thought of as a constructor for an Amount type.
func FloatAmount(float float64) Amount {
	return Amount(float * 100)
}

// DivideAmt allows Division by integers(Amounts).
func DivideAmt(top, bottom Amount) Amount {
	return (top*200 + bottom) / (bottom * 2)
}

// ToCurrency returns a string representation of a USD amount.
func (amt Amount) ToCurrency() string {
	str := strconv.Itoa(int(amt))

	b := bytes.NewBufferString(str)
	numCommas := (b.Len() - 2) / 3

	j := 0
	out := make([]byte, b.Len()+numCommas+2) // 2 extra placeholders for a `$` and a `.`
	for i, v := range b.Bytes() {
		if i == (b.Len() - 2) {
			out[j], _ = bytes.NewBufferString(".").ReadByte()
			j++
		} else if (i-1)%3 == 0 {
			out[j], _ = bytes.NewBufferString(",").ReadByte()
			j++
		} else if i == 0 {
			out[j], _ = bytes.NewBufferString("$").ReadByte()
			j++
		}
		out[j] = v
		j++
	}
	return string(out)
}

// ToVolume returns a string representation of a quantity or volume.
func (amt Amount) ToVolume() string {
	str := strconv.Itoa(int(amt))

	b := bytes.NewBufferString(str)
	numCommas := (b.Len() - 2) / 3

	j := 0
	out := make([]byte, b.Len()+numCommas+1) // 1 extra placeholders for a `.`
	for i, v := range b.Bytes() {
		if i == (b.Len() - 2) {
			out[j], _ = bytes.NewBufferString(".").ReadByte()
			j++
		} else if (i-1)%3 == 0 {
			out[j], _ = bytes.NewBufferString(",").ReadByte()
			j++
		}
		out[j] = v
		j++
	}
	return string(out)
}

// ToPercent returns a string representation of a percent.
func (amt Amount) ToPercent() string {
	str := strconv.Itoa(int(amt))

	b := bytes.NewBufferString(str)
	numCommas := (b.Len() - 2) / 3

	j := 0
	out := make([]byte, b.Len()+numCommas+2) // 1 extra placeholders for a `.`
	for i, v := range b.Bytes() {
		if i == (b.Len() - 2) {
			out[j], _ = bytes.NewBufferString(".").ReadByte()
			j++
		} else if (i-1)%3 == 0 {
			out[j], _ = bytes.NewBufferString(",").ReadByte()
			j++
		}
		out[j] = v
		j++
	}
	out[j], _ = bytes.NewBufferString("%").ReadByte()
	return string(out)
}

type datedMetric struct {
	Amount Amount    `json:"amount"`
	Date   time.Time `json:"date"`
}

func newAvg(lastAvg Amount, nTicks uint, tickAmt Amount) Amount {
	numerator := lastAvg*Amount(nTicks) + tickAmt
	return numerator / (Amount(nTicks) + 1)
}

func newMax(lastMax *datedMetric, newPrice Amount, timestamp time.Time) *datedMetric {
	if newPrice >= lastMax.Amount {
		return &datedMetric{Amount: newPrice, Date: timestamp}
	}
	return lastMax
}

func newMin(lastMin *datedMetric, newPrice Amount, timestamp time.Time) *datedMetric {
	if newPrice <= lastMin.Amount {
		return &datedMetric{Amount: newPrice, Date: timestamp}
	}
	return lastMin
}
