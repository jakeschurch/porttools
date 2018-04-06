package trading

import (
	"errors"
)

var (
	// ErrOrderNotValid indicates that an order is not valid, and should not be sent further in the pipeline.
	ErrOrderNotValid = errors.New("Order does not meet criteria as a valid order")
)
