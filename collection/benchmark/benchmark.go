package benchmark

import (
	"errors"
	"sync"

	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/instrument/security"
)

var (
	// ErrSecurityExists indicates that a security struct has already been allocated in an index
	ErrSecurityExists = errors.New("Security exists in index")

	// ErrNoSecurityExists indicates that a security struct has not been allocated in an index's securities map
	ErrNoSecurityExists = errors.New("Security does not exist in Securities map")
)

// NewIndex returns a new Index type; typically used for benchmarking a portfolio.
func NewIndex() *Index {
	index := Index{
		Securities: make(map[string]*security.Security),
	}
	return &index
}

// Index struct allow for the use of a benchmark to compare financial performance,
// Index could refer to one Security or many.
type Index struct {
	sync.RWMutex
	Securities map[string]*security.Security
}

// AddNew adds a new security to an Index's Securities map.
func (index *Index) AddNew(tick instrument.Tick) error {
	index.RLock()
	if _, exists := index.Securities[tick.Ticker]; exists {
		index.RUnlock()
		return ErrSecurityExists
	}
	index.RUnlock()
	index.Securities[tick.Ticker] = security.New(tick)
	index.Securities[tick.Ticker].UpdateMetrics(tick)
	return nil
}

// UpdateMetrics passes instrument.Tick to appropriate Security in securities map.
// Returns error if security not found in map.
func (index *Index) UpdateMetrics(tick instrument.Tick) error {
	index.RLock()
	_, exists := index.Securities[tick.Ticker]
	index.RUnlock()
	if !exists {
		return ErrNoSecurityExists
	}
	index.Securities[tick.Ticker].UpdateMetrics(tick)
	return nil
}
