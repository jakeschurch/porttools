package order

import (
	"time"

	"github.com/jakeschurch/porttools/instrument/holding"
	"github.com/jakeschurch/porttools/utils"
)

// Order struct hold information referring to the
// details of an execution of a financial asset transaction.
type Order struct {
	// it's either a buy or sell
	Buy    bool
	Status Status
	Logic  Logic
	Ticker string
	// NOTE: turn price + datetime into LastBid & LastAsk
	Bid, Ask, Volume utils.Amount
	Datetime         time.Time
}

// New returns a new order that will execute at nearest price.
func New(buy bool, ticker string, bid, ask, volume utils.Amount, datetime time.Time) *Order {
	return &Order{
		Buy:      buy,
		Status:   open,
		Logic:    market,
		Ticker:   ticker,
		Bid:      bid,
		Ask:      ask,
		Volume:   volume,
		Datetime: datetime,
	}
}

func (order *Order) ToPosition() *holding.Holding {
	bid := &utils.DatedMetric{Amount: order.Bid, Date: order.Datetime}
	ask := &utils.DatedMetric{Amount: order.Ask, Date: order.Datetime}

	return &holding.Holding{
		Ticker:   order.Ticker,
		Volume:   order.Volume,
		NumTicks: 1,
		LastBid:  bid, LastAsk: ask,
		AvgBid: bid.Amount, AvgAsk: ask.Amount,
		MaxBid: bid, MaxAsk: ask,
		MinBid: bid, MinAsk: ask,
		BuyPrice: ask,
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
