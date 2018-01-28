package porttools

import "errors"

// MarketPortfolio struct holds instances of finacial stock securities.
type MarketPortfolio struct {
	Instruments map[string]*Security
}

// Portfolio struct holds instances of finacial stock holdings.
type Portfolio struct {
	Instruments map[string]*Holding
	Orders      []*[]Order // reference to all holdings' order slices
	Benchmark   *Security
}

// AddtoPortfolio adds holding instrument to Portfolio instance.
func (p *Portfolio) AddtoPortfolio(h *Holding) (err error) {

	if _, exists := p.Instruments[h.Ticker]; !exists {
		p.Instruments[h.Ticker] = h
		p.Orders = append(p.Orders, &h.Orders)

		return nil
	}
	return errors.New("Holding already exists in Instruments map")

}

// AddtoPortfolio adds holding instrument to MarketPortfolio instance.
func (p *MarketPortfolio) AddtoPortfolio(s *Security) (err error) {
	if _, exists := p.Instruments[s.Ticker]; !exists {
		p.Instruments[s.Ticker] = s
		return nil
	}
	return errors.New("Security already exists in Instruments map")

}
