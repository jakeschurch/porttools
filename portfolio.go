// Package porttools allows for storage of a portfolio struct.
package porttools

// Portfolio struct holds attributes relative to a financial security portfolio.
type Portfolio struct {
	Securities []*Security
	Benchmark  *Security
}
