// Package portfolio allows for storage of a portfolio struct.
package portfolio

import s "porttools/security"

// Portfolio struct holds attributes relative to a financial security portfolio.
type Portfolio struct {
	Securities []*s.Security
	Benchmark  *s.Security
}
