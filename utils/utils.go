package utils

import (
	"bytes"
	"errors"
	"strconv"
	"time"
)

var (
	// ErrEmptySlice indicates a slice with 0 elements
	ErrEmptySlice = errors.New("Slice has 0 elements")
)

// CostMethod regards the type of accounting management rule
// is implemented for selling securities.
type CostMethod int

const (
	// Lifo is for last-in-first-out
	Lifo CostMethod = iota - 1
	// Fifo is for first-in-first-out
	Fifo
)

// Amount is a representation of fractional volumes. To get around floating-point erroneous behavior, multiply volume by 100 and cap it as an integer.
type Amount int64

// FloatAmount converts a float64 value to an amount type. Can be thought of as a constructor for an Amount type.
func FloatAmount(float float64) Amount {
	return Amount(float * 100)
}

// DivideAmt allows Division by integers(Amounts).
func DivideAmt(top, bottom Amount) Amount {
	return (top*200 + bottom) / (bottom * 2)
}

// ToCurrency returns a string representation of a USD amount.
func (amt Amount) ToCurrency() string {
	str := strconv.Itoa(int(amt))

	b := bytes.NewBufferString(str)

	numCommas := (b.Len() - 2) / 3

	j := 0
	out := make([]byte, b.Len()+numCommas+3) // 2 extra placeholders for a `$` and a `.`

	for i, v := range b.Bytes() {
		if i == (b.Len() - 2) {
			out[j], _ = bytes.NewBufferString(".").ReadByte()
			j++
		} else if (i-1)%3 == 0 && b.Len() > 4 {
			out[j], _ = bytes.NewBufferString(",").ReadByte()
			j++
		} else if i == 0 {
			out[j], _ = bytes.NewBufferString("$").ReadByte()
			j++
		}
		out[j] = v
		j++
	}
	return string(out)
}

func (amt Amount) String() string {
	return string(amt)
}

// ToVolume returns a string representation of a quantity or volume.
func (amt Amount) ToVolume() string {
	str := strconv.Itoa(int(amt))

	b := bytes.NewBufferString(str)
	numCommas := (b.Len() - 2) / 3

	j := 0
	out := make([]byte, b.Len()+numCommas+1) // 1 extra placeholders for a `.`
	for i, v := range b.Bytes() {
		if i == (b.Len() - 2) {
			out[j], _ = bytes.NewBufferString(".").ReadByte()
			j++
		} else if (i-1)%3 == 0 {
			out[j], _ = bytes.NewBufferString(",").ReadByte()
			j++
		}
		out[j] = v
		j++
	}
	return string(out)
}

// ToPercent returns a string representation of a percent.
func (amt Amount) ToPercent() string {
	str := strconv.Itoa(int(amt))

	b := bytes.NewBufferString(str)
	numCommas := (b.Len() - 2) / 3

	j := 0
	out := make([]byte, b.Len()+numCommas+2) // 1 extra placeholders for a `.`
	for i, v := range b.Bytes() {
		if i == (b.Len() - 2) {
			out[j], _ = bytes.NewBufferString(".").ReadByte()
			j++
		} else if (i-1)%3 == 0 {
			out[j], _ = bytes.NewBufferString(",").ReadByte()
			j++
		}
		out[j] = v
		j++
	}
	out[j], _ = bytes.NewBufferString("%").ReadByte()
	return string(out)
}

// DatedMetric ...TODO
type DatedMetric struct {
	Amount Amount    `json:"amount"`
	Date   time.Time `json:"date"`
}

// Avg ..TODO
func Avg(lastAvg Amount, nTicks uint, tickAmt Amount) Amount {
	numerator := lastAvg*Amount(nTicks) + tickAmt
	return numerator / (Amount(nTicks) + 1)
}

// Max .. TODO
func Max(lastMax *DatedMetric, newPrice Amount, timestamp time.Time) *DatedMetric {
	if newPrice >= lastMax.Amount {
		return &DatedMetric{Amount: newPrice, Date: timestamp}
	}
	return lastMax
}

// Min ..TODO
func Min(lastMin *DatedMetric, newPrice Amount, timestamp time.Time) *DatedMetric {
	if newPrice <= lastMin.Amount {
		return &DatedMetric{Amount: newPrice, Date: timestamp}
	}
	return lastMin
}
