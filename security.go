// Package security allows for storage of information regarding particular securities.
package security

import (
	"fmt"
	"time"
)

// Security struct holds attributes relative to a security,
// including ticker and historical information.
type Security struct {
	Ticker          string
	Quantity        float64
	HistData        []TickData
	AdditionalAttrs map[string]interface{}
}

// newTickData creates a new tick instance, and appends it
// to the tickData slice, HistData.
func (s *Security) newTickData(date time.Time, price float64, volume float64) {
	tick := TickData{date, price, volume}
	s.HistData = append(s.HistData, tick)
}

// TickData is a struct that should not be used on its own, and is aggregated
// in a Security's HistData slice.
type TickData struct {
	Date          time.Time
	Price, Volume float64
}

// Kwarg struct allows for additional/optional arguments in a go function.
type Kwarg struct {
	name  string
	value interface{}
}

//TODO: create test file

// Handler is an aggregation struct holding all active securities and attributes.
type Handler struct {
	Securities []Security
}

// newSecurity creates a new Security instance and appends it to the
// Handler's Securities slice.
func (h *Handler) newSecurity(ticker string) {
	//newSecurity := Security{ticker, nil}
	//h.Securities = append(h.Securities, newSecurity)
}

func main() {
	fmt.Println("hello")
}
