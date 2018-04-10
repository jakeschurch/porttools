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
		Holdings: collection.NewHoldingList(),
	}
	return &index
}

// Index struct allow for the use of a benchmark to compare financial performance,
// Index could refer to one Security or many.
type Index struct {
	mu       sync.RWMutex
	Holdings *collection.HoldingList
}

// Insert adds a new holding to an Index's holdings list.
func (index *Index) insert(h *instrument.Holding, q instrument.Quote) error {
	return index.Holdings.InsertUpdate(h, q)
}

// Update will use q to bring holding metrics up to date.
// If Holdings.Update returns an error (empty list), insert new holding.
func (index *Index) Update(q instrument.Quote) {
	if err := index.Holdings.Update(q); err != nil {
		index.Holdings.Insert(q)
	}
}
