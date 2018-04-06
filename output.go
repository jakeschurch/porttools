package porttools

import (
	"encoding/csv"
	"log"
	"os"
	"time"
)

// OutputFmt ...todo
type OutputFmt int

// Output specifies format of results wanted
const (
	CSV OutputFmt = iota
	// TODO: JSON
)

func positionHeaders() []string {

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
	output = append(output, positionHeaders())

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
	AvgVolume Amount

	BuyValue Amount
	EndValue Amount

	AvgBid Amount
	MaxBid *datedMetric
	MinBid *datedMetric

	AvgAsk Amount
	MaxAsk *datedMetric
	MinAsk *datedMetric
	// TODO REVIEW: avgReturn
	PctReturn Amount
	Alpha     Amount
}

func (result *result) update(pos Position) {
	result.AvgBid += pos.AvgBid
	result.AvgAsk += pos.AvgAsk

	result.BuyValue += pos.BuyPrice.Amount * pos.Volume
	result.EndValue += pos.SellPrice.Amount * pos.Volume

	result.Filled++

	result.MaxBid = newMax(result.MaxBid, pos.MaxBid.Amount, pos.MaxBid.Date)
	result.MinBid = newMin(result.MinBid, pos.MinBid.Amount, pos.MinBid.Date)

	result.MaxAsk = newMax(result.MaxAsk, pos.MaxAsk.Amount, pos.MaxAsk.Date)
	result.MinAsk = newMin(result.MinAsk, pos.MinAsk.Amount, pos.MinAsk.Date)
	return
}

func (result *result) averageize() {
	amtFilled := Amount(result.Filled)
	result.AvgBid = DivideAmt(result.AvgBid, amtFilled)
	result.AvgAsk = DivideAmt(result.AvgAsk, amtFilled)

	result.PctReturn = DivideAmt((result.EndValue - result.BuyValue), result.BuyValue)
	return
}
