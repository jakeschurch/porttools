// Package porttools allows for storage of information regarding particular securities.
package porttools

import (
	"bytes"
	"strconv"
	"time"
)

// ...modelable? investmentable?
// TODO: type marketable interface {
// 	updateMetrics()
// }

// Amount ... TODO
type Amount uint64

// Currency ... TODO
func (c Amount) Currency() string {
	str := strconv.Itoa(int(c))

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

// Security structs hold information regarding a financial asset for the entire
// life of the financial asset in a trading environment. Because a Security struct
// holds aggregate information regarding a financial asset, it is embeded into an Index or Benchmark.
type Security struct {
	Ticker                        string
	NumTicks                      uint
	AvgVolume, AvgPrice           Amount
	LastPrice, MaxPrice, MinPrice datedMetric
	MaxVolume, MinVolume          datedMetric
	BuyPrice, SellPrice           datedMetric
}

// NewSecurity instantiates a new security from Tick data.
func NewSecurity(tick Tick) *Security {
	firstPrice := datedMetric{Amount: tick.Price, Date: tick.Datetime}
	firstVolume := datedMetric{Amount: tick.Volume, Date: tick.Datetime}
	return &Security{
		Ticker: tick.Ticker, NumTicks: 1,
		LastPrice: firstPrice, BuyPrice: firstPrice,
		AvgPrice: tick.Price, AvgVolume: tick.Volume,
		MaxPrice: firstPrice, MinPrice: firstPrice,
		MaxVolume: firstVolume, MinVolume: firstVolume,
	}
}

// TODO updateMetrics ...
func (security *Security) updateMetrics(tick Tick) {

}

type datedMetric struct {
	Amount Amount
	Date   time.Time
}

// Position structs refer the holding of a financial asset.
type Position struct {
	Ticker                        string
	Volume                        Amount
	NumTicks                      uint
	AvgPrice                      Amount
	LastPrice, MaxPrice, MinPrice datedMetric
	BuyPrice, SellPrice           datedMetric
}

func (pos *Position) updateMetrics(tick Tick) (ok bool) {
	pos.LastPrice = datedMetric{tick.Price, tick.Datetime}

	pos.AvgPrice = func() Amount {
		numerator := (pos.AvgPrice * Amount(pos.NumTicks)) + tick.Price
		return numerator / (Amount(pos.NumTicks) + 1)
	}()
	pos.NumTicks++

	pos.MaxPrice = func() datedMetric {
		if tick.Price >= pos.MaxPrice.Amount {
			return datedMetric{Amount: tick.Price, Date: tick.Datetime}
		}
		return pos.MaxPrice
	}()

	pos.MinPrice = func() datedMetric {
		if tick.Price <= pos.MinPrice.Amount {
			return datedMetric{Amount: tick.Price, Date: tick.Datetime}
		}
		return pos.MinPrice
	}()
	return true
}

// Tick structs holds information about a financial asset at a specific point in time.
type Tick struct {
	Ticker   string
	Price    Amount
	Volume   Amount
	BidSize  Amount
	AskSize  Amount
	Datetime time.Time
}

// Kwarg struct allows for add'l args/attrs to a class or func.
// NOTE: is this really needed?
type Kwarg struct {
	name  string
	value interface{}
}
