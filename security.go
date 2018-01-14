// Package porttools allows for storage of information regarding particular securities.
package porttools

import (
	"errors"
	"time"
)

// Security struct holds attributes relative to a financial security,
// such as a stock ticker, as well as the tick data of the instrument.
type Security struct {
	Ticker          string
	HistData        []TickData
	AdditionalAttrs []Kwarg
}

// NewSecurity initializes a new Securty struct, and returns a references
// the the memory location of the newly created Security.
func NewSecurity(t string, kwargs []Kwarg) *Security {
	s := Security{Ticker: t, AdditionalAttrs: kwargs}

	return &s
}

// Holding struct is a subclass of the Security struct,
// allowing attributes regarding orders and the active quantity of
// shares held to be defined.
type Holding struct {
	*Security      // Anonymous field
	ActiveQuantity float64
	Orders         []Order
}

// NewHolding initializes a new Holding struct, and returns a reference
// to the memory location of the newly created Holding.
func NewHolding(s *Security, buyOrder Order) *Holding {
	var (
		orders = []Order{}
	)

	Holding := &Holding{s, 0, orders}
	Holding.Transact(buyOrder)

	return Holding
}

// Transact conducts agreement between Holding and Order
func (h *Holding) Transact(o Order) (err error) {

	if o.TransactionT == Sell && h.ActiveQuantity-o.Quantity >= 0 {
		h.ActiveQuantity -= o.Quantity
	} else if o.TransactionT == Buy {
		h.ActiveQuantity += o.Quantity
	} else {
		return errors.New("cannot hold less than 0 shares")
	}

	h.Orders = append(h.Orders, o) // Add order to the Holding's Orders slice

	return nil
}

// TickData is a struct that should not be used on its own, and is aggregated
// in a Holding's HistData slice.
// Whenever a TickData slice is instantiated - it should be stored in a
// Holding instance of HistData.
// TODO(TickData) Create Example of storing tickData in Holding instance
type TickData struct {
	Price, Volume, BidSize, AskSize float64
	Date                            time.Time // NOTE: Data Date format: HHMMSSxxxxxxxxx
}

// Order stores information regarding a stock transaciton.
type Order struct {
	TransactionT TransactionType
	OrderT       OrderType
	Price        float64
	Quantity     float64
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

// Handler is an aggregation struct holding all active securities.
type Handler struct {
	Securities []*Holding
}
