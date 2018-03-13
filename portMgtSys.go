package porttools

// TODO: portMgtSys ...
type portMgtSys struct {
	minPctCash     float32
	maxLot         Amount
	maxTradeRisk   float32
	profitFactor   float32
	MaxDrawdownCap Amount
}

// OrderSys ... TODO
type OrderSys struct {
	orderChan chan Order
}
