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

func (prfmLog *PrfmLog) getHeaders() []string {

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
		result.AvgVolume.ToVolume(),
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

func (prfmLog *PrfmLog) toCSV() (ok bool) {
	var output [][]string
	output = append(output, prfmLog.getHeaders())

	for _, result := range prfmLog.results {
		output = append(output, result.ToSlice())
	}

	outFile, fileErr := os.Create("~/Desktop/simOutput.csv")
	if fileErr != nil {
		log.Fatal("Cannot create file", fileErr)
	}

	// TEMP: allow filename as method argument
	w := csv.NewWriter(outFile)

	for _, row := range output {
		if err := w.Write(row); err != nil {
			log.Fatalln("Could not write row")
		}
	}
	return true
}

func (oms *OMS) outputResults() {
	// TODO: json, csv outs
	switch oms.prfmLog.outFmt {
	case CSV:
		oms.prfmLog.toCSV()
	}
}

type result struct {
	Ticker    string `json:"ticker"`
	Filled    uint   `json:"filled"`
	AvgVolume Amount `json:"avgVolume"`

	BuyValue Amount `json:"buyValue"`
	EndValue Amount `json:"endValue"`

	AvgBid Amount      `json:"avgBid"`
	MaxBid datedMetric `json:"maxBid"`
	MinBid datedMetric `json:"minBid"`

	AvgAsk Amount      `json:"avgAsk"`
	MaxAsk datedMetric `json:"maxAsk"`
	MinAsk datedMetric `json:"minAsk"`
	// TODO REVIEW: avgReturn
	PctReturn Amount `json:"pctReturn"`
	Alpha     Amount `json:"alpha"`
}

func (result *result) update(pos *Position) {
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

func (result *result) averageize() { // REVIEW:  oh dear god
	amtFilled := Amount(result.Filled)
	result.AvgBid /= amtFilled
	result.AvgAsk /= amtFilled

	result.PctReturn = (result.EndValue - result.BuyValue) / result.BuyValue
	return
}
