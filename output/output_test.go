package output

import (
	"testing"

	"github.com/jakeschurch/porttools/collection/benchmark"
	"github.com/jakeschurch/porttools/instrument"
)

// FIX: UPDATE/REALLOCATE

func TestGetResults(t *testing.T) {
	type args struct {
		closedholdings []instrument.Holding
		benchmark      *benchmark.Index
		outputFormat   Format
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetResults(tt.args.closedholdings, tt.args.benchmark, tt.args.outputFormat)
		})
	}
}
