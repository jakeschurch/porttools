package order

import (
	"github.com/jakeschurch/porttools/instrument"
)

// Order struct hold information referring to the
// details of an execution of a financial asset transaction.
type Order struct {
	instrument.Quote

	// it's either a buy or sell
	Buy    bool
	Status Status
	Logic  Logic
}

// New returns a new order that will execute at nearest price.
func New(buy bool, q instrument.Quote) *Order {
	return &Order{
		Quote:  q,
		Buy:    buy,
		Status: open,
		Logic:  market, // TEMP: for now, only accepting market orders
	}
}

func (o Order) GetUnderlying() instrument.Financial {
	return o.Quote
}

func (o Order) Ticker() string {
	return o.Ticker()
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
