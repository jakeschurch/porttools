package porttools

import (
	"log"
	"sync"
	"time"
)

// TODO:  rebalance routine for entry/exit orders
func newOMS(cashAmt Amount, costMethod CostMethod, toIgnore []string, outFmt OutputFmt) *OMS {
	oms := OMS{
		Port:       NewPortfolio(cashAmt),
		benchmark:  newIndex(),
		openOrders: make([]*Order, 0),
		strat:      newStrategy(toIgnore, costMethod),
		prfmLog:    newPrfmLog(outFmt),
		// create channels
		orderChan: make(chan *Order, 1024),
		benchChan: make(chan *Tick, 1024),
		portChan:  make(chan *Tick, 1024),
		tickChan:  make(chan *Tick),
		closing:   make(chan struct{}, 1),
	}
	go oms.mux()
	return &oms

}

// OMS acts as an `Order Management System` to test trading signals and fill orders.
type OMS struct {
	Port       *Portfolio
	benchmark  *Index
	openOrders []*Order
	strat      *strategy
	prfmLog    *PrfmLog
	// Channels
	orderChan chan *Order
	tickChan  chan *Tick
	portChan  chan *Tick
	benchChan chan *Tick
	openChan  chan *Order
	closing   chan struct{}
}

// mux is the function that allows for OMS to integrate as a part of the simulation pipeline.
func (oms *OMS) mux() {
	for oms.orderChan != nil ||
		oms.openChan != nil ||
		oms.benchChan != nil { // oms.tickChan != nil
		select {
		// case tick, ok := <-oms.tickChan:
		// 	if !ok {
		// 		log.Println("OMS's tick channel has been closed.")
		// 		oms.tickChan = nil
		// 		continue
		// 	}
		// 	go oms.processTick(tick)
		case tick, ok := <-oms.benchChan:
			if !ok {
				log.Println("OMS's order channel has been closed.")
				oms.benchChan = nil
				continue
			}
			oms.benchmark.updateSecurity(*tick)

		case tick, ok := <-oms.portChan:
			if !ok {
				log.Println("OMS's order channel has been closed.")
				oms.portChan = nil
				continue
			}
			oms.Port.updatePositions(tick)

		case order, ok := <-oms.orderChan:
			if !ok {
				log.Println("OMS's order channel has been closed.")
				oms.orderChan = nil
				continue
			}
			// REVIEW: use go func?
			oms.Transact(order)

		case openOrder, ok := <-oms.openChan:
			if !ok {
				log.Println("OMS's open order channel has been closed.")
				oms.openChan = nil
				continue
			}
			oms.openOrders = append(oms.openOrders, openOrder)
			// default:
			// 	time.Sleep(1 * time.Millisecond)
		}
	}
	// REVIEW TEMP: where should these actually go?
	close(oms.tickChan)
	close(oms.orderChan)

}
func (oms *OMS) closeHandle() {
	oms.getResults()

	oms.prfmLog.quit()
	oms.closing <- struct{}{}
}

func (oms *OMS) closeOrders() {
	if len(oms.openOrders) == 0 {
		return
	}
	openOrder := oms.openOrders[0]
	newOrder := &Order{
		Buy:      false,
		Status:   open,
		Logic:    market,
		Ticker:   openOrder.Ticker,
		Volume:   openOrder.Volume,
		Price:    oms.Port.Active[openOrder.Ticker].positions[0].LastBid.Amount,
		Datetime: oms.Port.Active[openOrder.Ticker].positions[0].LastBid.Date,
	}
	go func(order *Order) { oms.orderChan <- order }(newOrder)

	// reduce size of openOrders slice since the 0th element has been closed.
	oms.openOrders = oms.openOrders[0:]
	oms.closeOrders()

	return
}

func (oms *OMS) processTick(tick *Tick) {
	if _, exists := oms.strat.ignore[tick.Ticker]; !exists {

		oms.benchChan <- tick

		if order, signal := oms.strat.algo.EntryLogic(tick); signal &&
			oms.strat.algo.ValidOrder(oms, order) {

			oms.orderChan <- order
			log.Println("buy order has been sent!")
		}

		if slice := oms.existsInOrders(tick.Ticker); len(slice) > 0 {
			var wg sync.WaitGroup

			for _, slicedOrder := range slice {
				wg.Add(1)
				go func(openOrder *Order) {
					if order, signal := oms.strat.algo.ExitLogic(tick, openOrder); signal &&
						oms.strat.algo.ValidOrder(oms, order) {

						oms.orderChan <- order
						log.Println("sell order has been sent")
					}
				}(slicedOrder)
				wg.Wait()
			}
			oms.portChan <- tick
		}
	}
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

// NewMarketOrder returns a buy order that will execute at nearest price.
// TEMP(NewMarketOrder) TODO
func NewMarketOrder(buy bool, ticker string, price, volume Amount, datetime time.Time) *Order {
	return &Order{
		Buy:      buy,
		Status:   open,
		Logic:    market,
		Ticker:   ticker,
		Price:    price,
		Volume:   volume,
		Datetime: datetime,
	}
}

// Order struct hold information referring to the
// details of an execution of a financial asset transaction.
type Order struct {
	// it's either buy or sell
	Buy    bool
	Status OrderStatus
	Logic  TradeLogic
	Ticker string
	// NOTE: turn price + datetime into LastBid & LastAsk
	Price    Amount
	Volume   Amount
	Datetime time.Time
}

func (order *Order) toPosition(volume Amount) *Position {
	// TEMP: when have time - flush this out fully
	return &Position{
		Ticker: order.Ticker, Volume: volume,
		BuyPrice: datedMetric{order.Price, order.Datetime},
		NumTicks: 0,
	}
}

// OrderStatus variables refer to a status of an order's execution.
type OrderStatus int

const (
	open OrderStatus = iota // 0
	closed
	cancelled
	expired // 3
)

// TradeLogic is used to identify when the order should be executed.
type TradeLogic int

const (
	market TradeLogic = iota // 0
	limit
	stopLimit
	stopLoss
	dayTrade // 4
)
