package porttools

func newBacktestEngine(cashAmt Amount, costMethod CostMethod, toIgnore []string) *BacktestEngine {
	btEngine := &BacktestEngine{
		OMS: newOMS(cashAmt, costMethod, toIgnore),
		// TODO: portLog
	}
	return btEngine
}

// BacktestEngine is the centralized struct that everything is occuring through within a simulation.
type BacktestEngine struct {
	Log *PerformanceLog
	OMS *OMS
}

// IDEA: send closed orders to PerformanceLog Closed slice instead of back to Portfolio's ClosedPosition Slice
// OPTIMIZE: instead of using sync.Mutexes, use channels/non-blocking functions

func newOMS(cashAmt Amount, costMethod CostMethod, toIgnore []string) *OMS {
	return &OMS{
		portfolio:  NewPortfolio(cashAmt),
		openOrders: make([]*Order, 0),
		strategy:   newStrategy(toIgnore, costMethod),
		tickChan:   make(chan *Tick),
		closing:    make(chan struct{}, 1),
	}
}

// OMS acts as an `Order Management System` to test trading signals and fill orders.
type OMS struct {
	portfolio  *Portfolio
	benchmark  *Index
	openOrders []*Order
	strategy   *strategy
	orderChan  chan *Order
	tickChan   chan *Tick
	closing    chan struct{}
}

func (oms *OMS) handle() {

	go func() {
		for {
			select {
			case order := <-oms.orderChan:
				oms.portfolio.Transact(order, oms.strategy.costMethod)
			case tick := <-oms.tickChan:
				go oms.processTick(tick)
			case <-oms.closing:
				return
			default:
			}
		}
	}()
	<-oms.closing
}

func (oms *OMS) existsInOrders(ticker string) []*Order {
	orders := make([]*Order, 0)

	for _, order := range oms.openOrders {
		if order.Ticker == ticker {
			orders = append(orders, order)
		}
	}
	return orders
}

func (oms *OMS) processTick(tick *Tick) {
	if _, exists := oms.strategy.ignore[tick.Ticker]; !exists {
		oms.benchmark.updateSecurity(tick)

		if order, signal := oms.strategy.algo.EntryLogic(tick); signal && oms.strategy.algo.ValidOrder(oms.portfolio, tick, order) {
			go func() { oms.orderChan <- order }()
		}
		if slice := oms.existsInOrders(tick.Ticker); len(slice) > 0 {

			for _, slicedOrder := range slice {
				go func(openOrder *Order) {
					if order, signal := oms.strategy.algo.ExitLogic(tick, openOrder); signal && oms.strategy.algo.ValidOrder(oms.portfolio, tick, openOrder) {
						go func() { oms.orderChan <- order }()

					}
				}(slicedOrder)
			}
			oms.portfolio.updatePositions(tick)
		}
	}
}

func (oms *OMS) closeHandle() {
	oms.closing <- struct{}{}
}

// newStrategy creates a new Strategy instance used in the backtesting process.
func newStrategy(toIgnore []string, costMethod CostMethod) *strategy {
	toIgnoreMap := make(map[string]bool)
	for _, ticker := range toIgnore {
		toIgnoreMap[ticker] = true
	}
	strat := &strategy{
		ignore:     toIgnoreMap,
		costMethod: costMethod,
	}
	return strat
}

// Algorithm is an interface that needs to be implemented in the pipeline by a user to fill orders based on the conditions that they specify.
type Algorithm interface {
	EntryLogic(*Tick) (*Order, bool)
	ExitLogic(*Tick, *Order) (*Order, bool)
	ValidOrder(*Portfolio, *Tick, *Order) bool
}

// strategy ... TODO
type strategy struct {
	algo       Algorithm
	costMethod CostMethod
	ignore     map[string]bool
}

// PerformanceLog conducts performance analysis.
type PerformanceLog struct {
	Closed PositionSlice
	orders Queue
	// benchmark *Index
}

// - max-drawdown
// - % profitable
// - total num trades
// - winning/losing trades
// - trading period length

// only see different trades -> simulate market that has traders doing different trades and what their aggregate position look like
