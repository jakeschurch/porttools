// Package porttools allows for storage of information regarding particular securities.
package porttools

import (
	"time"
)

// Security struct holds attributes relative to a financial security,
// such as a stock ticker, as well as the tick data of the instrument.
type Security struct {
	Ticker                          string
	numTicks                        int
	LastPrice, PriceMax, PriceMin   float64
	MaxVolume, MinVolume, AvgVolume float64
	AddlAttrs                       []Kwarg
}

// NewSecurity initializes a new Securty struct, and returns a references
// the the memory location of the newly created Security.
func NewSecurity(t string, kwargs []Kwarg) *Security {
	s := Security{Ticker: t, numTicks: 0, AddlAttrs: kwargs}

	return &s
}

// Position struct is a subclass of the Security struct,
// allowing attributes regarding orders and the active Volume of
// shares held to be defined.
type Position struct {
	Ticker                            string
	Volume                            float64
	numTicks                          int
	LastPrice, PriceBought, PriceSold float64
	PriceAvg, PriceMax, PriceMin      float64
	DateBought, DateSold              time.Time
}

func (pos *Position) sellShares(order *Order) *Position {
	pos.Volume = order.Volume
	pos.PriceSold = order.Price
	pos.DateSold = order.Date

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
	OrderT       OrderType
	Ticker       string
	Price        float64
	Volume       float64
	Date         time.Time
}

// OrderType used to identify type of order.
type OrderType int

const (
	market OrderType = iota // 0
	limit
	stopLimit
	stopLoss
	day
	open // 5
)

// TransactionType used to identify type of transaction.
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
