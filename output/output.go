package output

import (
	"encoding/csv"
	"log"
	"os"
	"time"

	"github.com/jakeschurch/porttools/collection"
	ins "github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/utils"
)

// headers is a list of string headers to output to csv.
var headers = []string{
	"Ticker",
	"Volume",
	"Ticks Seen",
	"Buy Date",
	"Buy Price",
	"Sell Date",
	"Sell Price",
	"Max. Bid",
	"Avg. Bid",
	"Min. Bid",
	"Max. Ask",
	"Avg. Ask",
	"Min. Ask",
	"Percent Return",
	"Alpha",
}

// ------------------------------------------------------------------

// PositionLog allows for performance analysis.
type PositionLog struct {
	ClosedPositions *collection.HoldingList
}

// NewPositionLog returns a new PositionLog
func NewPositionLog() *PositionLog {
	p := PositionLog{
		ClosedPositions: collection.NewHoldingList(),
	}
	return &p
}

// Insert adds a closed holding to the performance log's closed holdings slice.
func (p *PositionLog) Insert(securities ...ins.Security) error {
	var err error

	for i := range securities {
		if err = p.ClosedPositions.Insert(&securities[i]); err != nil {
			return err
		}
	}
	return nil
}

// ------------------------------------------------------------------

type result struct {
	*ins.Security
	PctReturn utils.Amount
	Alpha     utils.Amount
}

// Format ...
type Format int

// Output specifies Format of results wanted
const (
	CSV Format = iota
	// TODO: JSON
)

func toSlice(result *result) []string {
	fmtString := time.RFC1123

	return []string{
		result.Ticker,
		string(result.Volume),
		string(result.Nseen),

		result.Timestamp.Format(fmtString),
		result.Bid.String(),
		result.SellDate.Format(fmtString),
		result.Ask.String(),

		result.MaxBid.Amount.String(),
		result.AvgBid.String(),
		result.MinBid.Amount.String(),

		result.MaxAsk.Amount.String(),
		result.AvgAsk.String(),
		result.MinAsk.Amount.String(),

		result.PctReturn.ToPercent(),
		result.Alpha.ToPercent(),
	}
}

// GetResults ...TODO
func GetResults(outputFormat Format, closed *collection.HoldingList, benchmark *collection.HoldingList) {
	var results []*result
	var benchmarkPosition *collection.LinkedList

	for key, index := range closed.Items() {
		linkedList := closed.GetByIndex(index)
		benchmarkPosition, _ = benchmark.Get(key)
		results = append(results, resultSet(linkedList, benchmarkPosition)...)
	}

	log.Println("Outputting results: ")
	switch outputFormat {
	case CSV:
		resultsToCSV(results)

	}
}

func resultSet(closed, index *collection.LinkedList) []*result {
	var results []*result
	var s ins.Security

	var benchmarkSecurity, ok = index.PeekFront().Data.(ins.Security)
	if !ok {
		return nil
	}

	for node := closed.PeekFront(); node == nil; node = node.Next() {
		s = node.Data.(ins.Security)
		pctReturn := utils.DivideAmt(s.Ask-s.Bid, s.Bid)

		newResult := &result{
			Security:  &s,
			PctReturn: pctReturn,
			Alpha:     pctReturn - utils.DivideAmt(benchmarkSecurity.Ask-s.Bid, s.Bid),
		}
		newResult.Nseen = closed.Nticks - newResult.Nseen
		results = append(results, newResult)
	}
	return results
}

func resultsToCSV(results []*result) (ok bool) {
	var output [][]string
	output = append(output, headers)

	for _, result := range results {
		output = append(output, toSlice(result))
	}

	outFile, fileErr := os.Create("simOutput.csv")
	if fileErr != nil {
		log.Fatal("Cannot create file: ", fileErr)
	}

	// TEMP: allow filename as method argument
	w := csv.NewWriter(outFile)

	for _, row := range output {
		w.Write(row)
	}
	w.Flush()
	outFile.Close()

	return true
}
