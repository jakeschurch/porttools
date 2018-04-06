package porttools

import (
	"testing"
	"time"
)

var (
	oms = NewOMS()
)

func TestOMS_closeOutOrder(t *testing.T) {
	order := NewMarketOrder(true, "AAPL", FloatAmount(50.00), FloatAmount(51.00), 10, time.Time{})
	oms.openOrders = append(oms.openOrders, order)

	oms.closeOutOrder(order)
	if len(oms.openOrders) != 0 {
		t.Errorf("Expected 0 open orders, got %d open orders", len(oms.openOrders))
		t.Fail()
	}
}
