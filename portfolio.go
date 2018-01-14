package porttools

// Portfolio struct holds attributes relative to a financial security portfolio.
type Portfolio struct {
	Holdings  []*Holding
	Benchmark *Security
}
