package collection

import (
	"testing"
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

func TestPut(t *testing.T) {
	type args struct {
		l   *LookupCache
		key string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Base case", args{mockLookupCache([]string{}, []int16{}), "AAPL"}, false},
		{"Check for existing ticker",
			args{
				mockLookupCache([]string{"AAPL"}, []int16{}),
				"AAPL"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Put(tt.args.l, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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
