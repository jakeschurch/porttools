package output

import (
	"encoding/csv"
	"log"
	"os"
	"time"

	"github.com/jakeschurch/porttools/instrument/holding"
	"github.com/jakeschurch/porttools/instrument/security"
	"github.com/jakeschurch/porttools/utils"
)

// Fmt ...todo
type Fmt int

// Output specifies format of results wanted
const (
	CSV Fmt = iota
	// TODO: JSON
)

func headers() []string {

	return []string{
		"Ticker",
		"Filled",
		"AvgVolume",
		"BuyValue",
		"EndValue",
		"AvgBid",
		"MaxBid (amount)",
		"MaxBid (date)",
		"MinBid (amount)",
		"MinBid (date)",
		"AvgAsk",
		"MaxAsk (amount)",
		"MaxAsk (date)",
		"MinAsk (amount)",
		"MinAsk (date)",
		"PctReturn",
		"Alpha",
	}
}

func (result *result) ToSlice() []string {
	fmtString := time.RFC1123

	return []string{
		result.Ticker,
		string(result.Filled),
		result.AvgVolume.String(),
		result.BuyValue.ToCurrency(),
		result.EndValue.ToCurrency(),
		result.AvgBid.ToCurrency(),
		result.MaxBid.Amount.ToCurrency(),
		result.MaxBid.Date.Format(fmtString),
		result.MinBid.Amount.ToCurrency(),
		result.MinBid.Date.Format(fmtString),
		result.AvgAsk.ToCurrency(),
		result.MaxAsk.Amount.ToCurrency(),
		result.MaxAsk.Date.Format(fmtString),
		result.MinAsk.Amount.ToCurrency(),
		result.MinAsk.Date.Format(fmtString),
		result.PctReturn.ToPercent(),
		result.Alpha.ToPercent(),
	}
}

func resultsToCSV(results []result) (ok bool) {
	var output [][]string
	output = append(output, headers())

	for _, result := range results {
		output = append(output, result.ToSlice())
	}

	outFile, fileErr := os.Create("simOutput.csv")
	if fileErr != nil {
		log.Fatal("Cannot create file: ", fileErr)
	}

	// TEMP: allow filename as method argument
	w := csv.NewWriter(outFile)

	for _, row := range output {
		log.Println(row)
		w.Write(row)
	}
	w.Flush()
	outFile.Close()

	return true
}

type result struct {
	Ticker    string
	Filled    uint
	AvgVolume utils.Amount

	BuyValue utils.Amount
	EndValue utils.Amount

	AvgBid utils.Amount
	MaxBid *utils.DatedMetric
	MinBid *utils.DatedMetric

	AvgAsk utils.Amount
	MaxAsk *utils.DatedMetric
	MinAsk *utils.DatedMetric
	// TODO REVIEW: avgReturn
	PctReturn utils.Amount
	Alpha     utils.Amount
}

func (result *result) update(holding holding.Holding) {
	result.AvgBid += holding.AvgBid
	result.AvgAsk += holding.AvgAsk

	result.BuyValue += holding.BuyPrice.Amount * holding.Volume
	result.EndValue += holding.SellPrice.Amount * holding.Volume

	result.Filled++

	result.MaxBid = utils.Max(result.MaxBid, holding.MaxBid.Amount, holding.MaxBid.Date)
	result.MinBid = utils.Min(result.MinBid, holding.MinBid.Amount, holding.MinBid.Date)

	result.MaxAsk = utils.Max(result.MaxAsk, holding.MaxAsk.Amount, holding.MaxAsk.Date)
	result.MinAsk = utils.Min(result.MinAsk, holding.MinAsk.Amount, holding.MinAsk.Date)
	return
}

func (result *result) averageize() {
	amtFilled := utils.Amount(result.Filled)
	result.AvgBid = utils.DivideAmt(result.AvgBid, amtFilled)
	result.AvgAsk = utils.DivideAmt(result.AvgAsk, amtFilled)

	result.PctReturn = utils.DivideAmt((result.EndValue - result.BuyValue), result.BuyValue)
	return
}

// GetResults ...TODO
func GetResults(closedholdings []holding.Holding, benchmarkMap map[string]*security.Security, outputFormat Fmt) {
	var results []result

	// create slice of all holding keys
	holdingKeys := make([]string, 0)

	for _, holding := range closedholdings {

		if !findKey(holdingKeys, holding.Ticker) {
			holdingKeys = append(holdingKeys, holding.Ticker)
		}
	}
	for _, key := range holdingKeys {
		filtered := filter(closedholdings, key)
		results = append(results, createResult(filtered, benchmarkMap[key]))
	}

	log.Println("Outputting results: ")
	switch outputFormat {
	case CSV:
		resultsToCSV(results)
	}
}

func newResult(holding holding.Holding) result {
	result := result{
		Ticker:    holding.Ticker,
		Filled:    1,
		AvgVolume: holding.Volume,

		BuyValue: holding.BuyPrice.Amount * holding.Volume,
		EndValue: holding.SellPrice.Amount * holding.Volume,

		AvgBid: holding.AvgBid,
		MaxBid: holding.MaxBid,
		MinBid: holding.MinBid,

		AvgAsk: holding.AvgAsk,
		MaxAsk: holding.MaxAsk,
		MinAsk: holding.MinAsk,
	}
	return result
}
func createResult(holdings []holding.Holding, security *security.Security) result {
	var holdingResult result

	// loop through and update metrics accordingly
	for index, holding := range holdings {
		if index == 0 {
			holdingResult = newResult(holdings[index])
		}
		holdingResult.update(holding)
	}
	holdingResult.averageize()

	if security == nil {
		holdingResult.Alpha = utils.Amount(0)
		return holdingResult
	}
	// NOTE: this is NOT on an aggregate basis at the moment.
	benchmarkReturn := utils.DivideAmt((security.LastAsk.Amount - security.BuyPrice.Amount), security.BuyPrice.Amount)
	holdingResult.Alpha = holdingResult.PctReturn - benchmarkReturn
	return holdingResult
}

// selection sort for positions.
func selectionSort(A []*holding.Holding) []*holding.Holding {
	for i := 0; i < len(A)-1; i++ {
		min := i
		for j := i + 1; j < len(A); j++ {
			if A[j].Ticker < A[min].Ticker {
				min = j
			}
		}
		key := A[i]
		A[i] = A[min]
		A[min] = key
	}
	return A
}

func filter(positions []holding.Holding, key string) []holding.Holding {
	filtered := make([]holding.Holding, 0)

	for _, position := range positions {
		if position.Ticker == key {
			filtered = append(filtered, position)
		}
	}
	return filtered
}

func findKey(A []string, toFind string) bool {
	for _, key := range A {
		if key == toFind {
			return true
		}
	}
	return false
}
