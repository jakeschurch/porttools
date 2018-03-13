// Package porttools allows for storage of information regarding particular securities.
package porttools

// FIXME: BUG when changed tick struct from holding Price to AskPrice & BidPrice
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

// ParseFloat ... TODO
func FloatAmount(float float64) Amount {
	return Amount(float * 100)
}

// Currency ... TODO
func (amt Amount) Currency() string {
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

// Security structs hold information regarding a financial asset for the entire
// life of the financial asset in a trading environment. Because a Security struct
// holds aggregate information regarding a financial asset, it is embeded into an Index or Benchmark.
type Security struct {
	Ticker               string      `json:"ticker"`
	NumTicks             uint        `json:"numTicks"`
	AvgVolume            Amount      `json:"avgVolume"`
	AvgPrice             Amount      `json:"avgPrice"`
	LastPrice            datedMetric `json:"lastPrice"`
	MaxPrice             datedMetric `json:"maxPrice"`
	MinPrice             datedMetric `json:"minPrice"`
	MaxVolume, MinVolume datedMetric `json:"maxVolume"`
	BuyPrice, SellPrice  datedMetric `json:"buyPrice"`
}

// NewSecurity instantiates a new security from Tick data.
func NewSecurity(tick Tick) *Security {
	firstPrice := datedMetric{Amount: tick.Price, Date: tick.Timestamp}
	firstVolume := datedMetric{Amount: tick.Volume, Date: tick.Timestamp}
	return &Security{
		Ticker: tick.Ticker, NumTicks: 1,
		LastPrice: firstPrice, BuyPrice: firstPrice,
		AvgPrice: tick.Price, AvgVolume: tick.Volume,
		MaxPrice: firstPrice, MinPrice: firstPrice,
		MaxVolume: firstVolume, MinVolume: firstVolume,
	}
}

func newAvg(lastAvg Amount, nTicks uint, tickAmt Amount) Amount {
	numerator := lastAvg*Amount(nTicks) + tickAmt
	return numerator / (Amount(nTicks) + 1)
}

func (s *Security) updateMetrics(tick Tick) {
	func() {
		s.AvgPrice = newAvg(s.AvgPrice, s.NumTicks, tick.Price)
		s.AvgVolume = newAvg(s.AvgVolume, s.NumTicks, tick.Volume)
		s.LastPrice = datedMetric{tick.Price, tick.Timestamp}
		s.NumTicks++
	}()
	func() {
		if tick.Price >= s.MaxPrice.Amount {
			s.MaxPrice = datedMetric{Amount: tick.Price, Date: tick.Timestamp}
			return
		}
		if tick.Price <= s.MinPrice.Amount {
			s.MinPrice = datedMetric{Amount: tick.Price, Date: tick.Timestamp}
		}
	}()
	func() {
		if tick.Volume >= s.MaxVolume.Amount {
			s.MaxVolume = datedMetric{Amount: tick.Volume, Date: tick.Timestamp}
			return
		}
		if tick.Volume <= s.MinVolume.Amount {
			s.MinVolume = datedMetric{Amount: tick.Volume, Date: tick.Timestamp}
		}
	}()
}

type datedMetric struct {
	Amount Amount    `json:"amount"`
	Date   time.Time `json:"date"`
}

// Position structs refer the holding of a financial asset.
type Position struct {
	Ticker                        string      `json:"ticker"`
	Volume                        Amount      `json:"volume"`
	NumTicks                      uint        `json:"numTicks"`
	AvgPrice                      Amount      `json:"avgPrice"`
	LastPrice, MaxPrice, MinPrice datedMetric `json:"lastPrice"`
	BuyPrice, SellPrice           datedMetric `json:"buyPrice"`
}

func (pos *Position) updateMetrics(tick Tick) {
	go func() {
		pos.AvgPrice = newAvg(pos.AvgPrice, pos.NumTicks, tick.Price)
		pos.LastPrice = datedMetric{tick.Price, tick.Timestamp}
		pos.NumTicks++
	}()
	go func() {
		if tick.Price >= pos.MaxPrice.Amount {
			pos.MaxPrice = datedMetric{Amount: tick.Price, Date: tick.Timestamp}
			return
		}
		if tick.Price <= pos.MinPrice.Amount {
			pos.MinPrice = datedMetric{Amount: tick.Price, Date: tick.Timestamp}
		}
	}()
}

// Tick structs holds information about a financial asset at a specific point in time.
type Tick struct {
	Ticker    string    `json:"ticker"`
	Volume    Amount    `json:"volume"`
	BidPrice  Amount    `json:"bidPrice"`
	BidSize   Amount    `json:"bidSize"`
	AskPrice  Amount    `json:"askPrice"`
	AskSize   Amount    `json:"askSize"`
	Timestamp time.Time `json:"timestamp"`
}
