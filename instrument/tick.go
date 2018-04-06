package instrument

import (
	"time"

	"github.com/jakeschurch/porttools/utils"
)

// Tick structs holds information about a financial asset at a specific point in time.
type Tick struct {
	Ticker           string
	Bid, Ask         utils.Amount
	BidSize, AskSize utils.Amount
	Timestamp        time.Time
}
