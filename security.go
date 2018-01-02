// Package security allows for storage of information regarding particular securities.
package security

import (
	"errors"
	"time"
)

// Security struct holds attributes relative to a security,
// including ticker and historical information.
type Security struct {
	Ticker          string
	Quantity        float64
	HistData        []*TickData
	Orders          []Order
	AdditionalAttrs []*Kwarg
}

// Transact conducts agreement between Security and Order
func (s *Security) Transact(o Order) error {

	if o.TransactionT == Sell && s.Quantity-o.Quantity >= 0 {
		s.Quantity -= o.Quantity
	} else if o.TransactionT == Buy {
		s.Quantity += o.Quantity
	} else {
		return errors.New("cannot hold less than 0 shares")
	}

	s.Orders = append(s.Orders, o) // Add order to Orders slice

	return nil
}

// TickData is a struct that should not be used on its own, and is aggregated
// in a Security's HistData slice.
// Whenever a TickData slice is instantiated - it should be stored in a
// Security instance of HistData.
// TODO: Create Example of storing tickData in Security instance
type TickData struct {
	Price, Volume, BidSize, AskSize float64
	Date                            time.Time // NOTE: Data Date format: HHMMSSxxxxxxxxx
}

// Order stores information regarding a stock transaciton.
type Order struct {
	OrderT       OrderType
	TransactionT TransactionType
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
	Securities []*Security
}
