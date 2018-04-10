package utils

import "testing"

func TestAmount_ToCurrency(t *testing.T) {
	tests := []struct {
		name string
		amt  Amount
		want string
	}{
		{"1", FloatAmount(50.00), "$50.00"},
		{"2", FloatAmount(500.00), "$500.00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.amt.ToCurrency(); got != tt.want {
				t.Errorf("Amount.ToCurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}
