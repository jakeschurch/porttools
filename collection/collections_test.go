package collection

import (
	"testing"
	"time"

	"github.com/jakeschurch/porttools/instrument"
	"github.com/jakeschurch/porttools/utils"
)

func mockLookupCache(items []string, openSlots []int16) *LookupCache {
	l := &LookupCache{
		items:     make(map[string]int16),
		openSlots: append(make([]int16, 0), openSlots...),
		last:      -1,
	}

	for i := 0; i < len(items); i++ {
		Put(l, items[i])
	}

	return l
}

func TestGet(t *testing.T) {
	type args struct {
		l   *LookupCache
		key string
	}
	tests := []struct {
		name string
		args args
		want int16
	}{
		{"Base Functionality",
			args{mockLookupCache([]string{"AAPL"}, []int16{}), "AAPL"}, 0},
		{"Check that string does not exist",
			args{mockLookupCache([]string{"AAPL"}, []int16{}), "BABA"}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Get(tt.args.l, tt.args.key); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPut(t *testing.T) {
	type args struct {
		l   *LookupCache
		key string
	}
	tests := []struct {
		name      string
		args      args
		wantValue int16
		wantErr   bool
	}{
		{"Base case", args{mockLookupCache([]string{}, []int16{}), "AAPL"}, 0, false},
		{"Check for existing ticker",
			args{
				mockLookupCache([]string{"AAPL"}, []int16{}),
				"AAPL"}, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, err := Put(tt.args.l, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Put() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("Put() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func mockHolding() *instrument.Holding {
	ask := utils.FloatAmount(50.00)
	askDatedMetric := &utils.DatedMetric{Amount: ask, Date: time.Time{}}

	i := &instrument.Holding{
		Instrument: instrument.NewInstrument("GOOGL", 10),
		BuyPrice:   askDatedMetric,
	}
	return i
}

func mockTick() instrument.Tick {
	ask := utils.FloatAmount(50.00)
	bid := utils.FloatAmount(49.50)
	bidSz := utils.Amount(10)
	askSz := utils.Amount(10)

	return instrument.Tick{
		Ticker: "GOOGL",
		Bid:    bid, Ask: ask,
		BidSize: bidSz, AskSize: askSz,
		Timestamp: time.Time{}}
}

func mockLinkedList() *LinkedList {
	ask := utils.FloatAmount(50.00)
	askDatedMetric := &utils.DatedMetric{Amount: ask, Date: time.Time{}}
	bid := utils.FloatAmount(49.00)
	bidDatedMetric := &utils.DatedMetric{Amount: bid, Date: time.Time{}}

	return NewLinkedList(
		instrument.Asset{
			Quote:   mockHolding().Instrument,
			LastAsk: askDatedMetric, LastBid: bidDatedMetric,
			MaxBid: bidDatedMetric, MaxAsk: askDatedMetric,
			MinAsk: askDatedMetric, MinBid: bidDatedMetric})
}

func TestLinkedList_Push(t *testing.T) {
	type fields struct {
		Asset *instrument.Asset
		head  *LinkedNode
		tail  *LinkedNode
	}
	// Setup fields
	mockedList := mockLinkedList()
	testFields := fields{
		Asset: mockedList.Asset,
		head:  mockedList.head,
		tail:  mockedList.tail}

	type args struct {
		node *LinkedNode
		t    instrument.Tick
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Base case", testFields,
			args{node: NewLinkedNode(mockHolding())}},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LinkedList{
				Asset: tt.fields.Asset,
				head:  tt.fields.head,
				tail:  tt.fields.tail,
			}
			l.Push(tt.args.node)
			if tt.name == "Base case" {
				vol := l.Volume(0)
				if vol != 20 {
					t.Errorf("Error of List Volume. Got = %d, expected %d", vol, 20)
				}
			}
		})
	}
}
