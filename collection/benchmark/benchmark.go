package benchmark

import (
	"errors"
	"sync"

	"github.com/jakeschurch/porttools/collection"
	"github.com/jakeschurch/porttools/instrument"
)

var (
	// ErrSecurityExists indicates that a security struct has already been allocated in an index
	ErrSecurityExists = errors.New("Security exists in index")

	// ErrNoSecurityExists indicates that a security struct has not been allocated in an index's holdings map
	ErrNoSecurityExists = errors.New("Security does not exist in Securities map")
)

// NewIndex returns a new Index type; typically used for benchmarking a portfolio.
func NewIndex() *Index {
	index := Index{
		holdings: collection.NewHoldingList(),
	}
	return &index
}

// Index struct allow for the use of a benchmark to compare financial performance,
// Index could refer to one Security or many.
type Index struct {
	mu       sync.RWMutex
	holdings *collection.HoldingList
}

// Insert adds a new holding to an Index's holdings list.
func (index *Index) Insert(h *instrument.Holding, tick instrument.Tick) error {
	return index.holdings.InsertUpdate(h, tick)
}
