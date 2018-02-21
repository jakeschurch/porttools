// Package porttools allows for storage of information regarding particular securities.
package porttools

import (
	"time"
)

// NewSecurity ...TODO
func NewSecurity(tick Tick) (newSecurity *Security) {
	firstPrice := datedMetric{Amount: tick.Price, Date: tick.Datetime}
	firstVolume := datedMetric{Amount: tick.Volume, Date: tick.Datetime}
	return &Security{
		Ticker: tick.Ticker, NumTicks: 1,
		LastPrice: firstPrice, BuyPrice: firstPrice,
		AvgPrice: tick.Price, AvgVolume: tick.Volume,
		MaxPrice: firstPrice, MinPrice: firstPrice,
		MaxVolume: firstVolume, MinVolume: firstVolume}
}

// Security ...TODO
type Security struct {
	Ticker                        string
	NumTicks                      int
	LastPrice, MaxPrice, MinPrice datedMetric
	MaxVolume, MinVolume          datedMetric
	BuyPrice, SellPrice           datedMetric
	AvgPrice, AvgVolume           float64
}

func (security *Security) update(tick Tick) {
	// TODO
}

type datedMetric struct {
	Amount float64
	Date   time.Time
}

// Position ...TODO
type Position struct {
	Ticker                        string
	Volume                        float64
	NumTicks                      int
	AvgPrice                      float64
	LastPrice, MaxPrice, MinPrice datedMetric
	BuyPrice, SellPrice           datedMetric
}

func (pos *Position) update(tick Tick) {
	pos.LastPrice = datedMetric{tick.Price, tick.Datetime}

	pos.AvgPrice = func(pos *Position, tick Tick) float64 {
		s := pos.AvgPrice*float64(pos.NumTicks) + tick.Price
		return s / float64(pos.NumTicks+1)
	}(pos, tick)

	pos.NumTicks++

}

func (pos *Position) updateMetrics(tick Tick) (ok bool) {
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

func (pos *Position) sellShares(order *Order) *Position {
	pos.Volume = order.Volume
	pos.SellPrice = datedMetric{Amount: order.Price, Date: order.Date}

	return pos
}

// Tick ...TODO
type Tick struct {
	Ticker   string
	Price    float64
	Volume   float64
	BidSize  float64
	AskSize  float64
	Datetime time.Time
}

// Order stores information regarding a stock transactiton.
type Order struct {
	TransactionT TransactionType
	ExecutionT   ExecutionType
	Ticker       string
	Price        float64
	Volume       float64
	Date         time.Time
}

// ExecutionType is used to identify type of order.
type ExecutionType int

const (
	market ExecutionType = iota // 0
	limit
	stopLimit
	stopLoss
	day
	open // 5
)

// TransactionType used to identify either a buy or a sell.
type TransactionType int

// Implement TransactionType enum.
const (
	Buy TransactionType = iota
	Sell
)

// Kwarg struct allows for add'l args/attrs to a class or func.
type Kwarg struct {
	name  string
	value interface{}
}
