// Package porttools allows for storage of information regarding particular securities.
package porttools

import (
	"time"
)

// Security struct holds attributes relative to a financial security,
// such as a stock ticker, as well as the tick data of the instrument.
type Security struct {
	Ticker                        string
	NumTicks                      int
	LastPrice, MaxPrice, MinPrice datedMetric
	MaxVolume, MinVolume          datedMetric
	PriceBought, SoldAt           datedMetric
	AvgPrice, AvgVolume           float64
	AddlAttrs                     []Kwarg
}

type datedMetric struct {
	Amount float64
	Date   time.Time
}

// NewSecurity initializes a new Securty struct, and returns a references
// the the memory location of the newly created Security.
func NewSecurity(t string, kwargs []Kwarg) *Security {
	s := Security{Ticker: t, NumTicks: 0, AddlAttrs: kwargs}

	return &s
}

// Position struct is a subclass of the Security struct,
// allowing attributes regarding orders and the active Volume of
// shares held to be defined.
type Position struct {
	Ticker              string
	Volume              float64
	NumTicks            int
	LastPrice, AvgPrice float64
	MaxPrice, MinPrice  datedMetric
	BoughtAt, SoldAt    datedMetric
}

func (pos *Position) update(tick Tick) {
	pos.LastPrice = tick.Price

	pos.AvgPrice = func(pos *Position, tick Tick) float64 {
		s := pos.AvgPrice*float64(pos.NumTicks) + tick.Price
		return s / float64(pos.NumTicks+1)
	}(pos, tick)

	pos.NumTicks++

}

// func (pos *Position) updateMetric(tick Tick, p positionMetric) {
// 	switch p {
// 	case positionMetric.maxPrice:
// 		pass
// 	}
// }

func (pos *Position) updateMetric(tick Tick) (ok bool) {
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

type positionMetric int

const (
	maxPrice positionMetric = iota
	minPrice
	maxVolume
	minVolume
)

// func (pos *Position) checkMetrics(tick Tick) {
// 	switch tick.Price {
// 	case >= pos.MaxPrice.Amount:
// 		pos.MaxPrice = datedMetric{Amount: tick.Price, Date: tick.Datetime}
//
// 	case <= pos.MinPrice.Amount:
// 		pos.MinPrice = datedMetric{Amount: tick.Price, Date: tick.Datetime}
//
// 	}
// }

func (pos *Position) sellShares(order *Order) *Position {
	pos.Volume = order.Volume
	pos.SoldAt = datedMetric{Amount: order.Price, Date: order.Date}

	return pos
}

// NewPosition initializes a new Position struct, and returns a reference
// to the memory location of the newly created Position.
func NewPosition(security *Security) *Position {
	position := Position{Ticker: security.Ticker}
	return &position
}

// Tick is a struct that should not be used on its own, and is aggregated
// in a Position's HistData slice.
// Whenever a TickData slice is instantiated - it should be stored in a
// Position instance of HistData.
// TODO(Tick) Create Example of storing tickData in Position instance
type Tick struct {
	Ticker   string
	Price    float64
	Volume   float64
	BidSize  float64
	AskSize  float64
	Datetime time.Time
}

// Order stores information regarding a stock transaciton.
type Order struct {
	TransactionT TransactionType
	ExecutionT   ExecutionType
	Ticker       string
	Price        float64
	Volume       float64
	Date         time.Time
}

// ExecutionType used to identify type of order.
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
