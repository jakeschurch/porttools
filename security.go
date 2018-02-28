// Package porttools allows for storage of information regarding particular securities.
package porttools

import (
	"time"
)

// NewSecurity instantiates a new security from Tick data.
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

// Security structs hold information regarding a financial asset for the entire
// life of the financial asset in a trading environment. Because a Security struct
// holds aggregate information regarding a financial asset, it is embeded into an Index or Benchmark.
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

// Position structs refer the holding of a financial asset.
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

}

func (pos *Position) updateMetrics(tick Tick) (ok bool) {
	pos.AvgPrice = func() float64 {
		numerator := pos.AvgPrice*float64(pos.NumTicks) + tick.Price
		return numerator / float64(pos.NumTicks+1)
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

func (pos *Position) sellShares(order *Order, amtToSell float64) *Position {

	soldPos := func() *Position {
		posSold := *pos
		posSold.Volume = amtToSell
		posSold.SellPrice = datedMetric{order.Price, order.Datetime}
		return &posSold
	}()
	// Update active volume for pos
	pos.Volume = pos.Volume - amtToSell

	return soldPos
}

// Tick structs holds information about a financial asset at a specific point in time.
type Tick struct {
	Ticker   string
	Price    float64
	Volume   float64
	BidSize  float64
	AskSize  float64
	Datetime time.Time
}

// Order structs hold information referring to the details of the execution of a financial asset transaction.
type Order struct {
	// it's either buy or sell
	Buy      bool
	Status   OrderStatus
	Logic    TradeLogic
	Ticker   string
	Price    float64
	Volume   float64
	Datetime time.Time
}

// OrderStatus variables refer to a status of an order's execution.
type OrderStatus int

const (
	open OrderStatus = iota // 0
	completed
	canceled
	expired // 3
)

// TradeLogic is used to identify when the order should be executed.
type TradeLogic int

const (
	market TradeLogic = iota // 0
	limit
	stopLimit
	stopLoss
	day // 4
)

// Kwarg struct allows for add'l args/attrs to a class or func.
type Kwarg struct {
	name  string
	value interface{}
}
