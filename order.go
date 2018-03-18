package porttools

import (
	"log"
	"time"
)

func newOMS(cashAmt Amount, costMethod CostMethod, toIgnore []string, outFmt OutputFmt) *OMS {
	return &OMS{
		port:       NewPortfolio(cashAmt),
		benchmark:  newIndex(),
		openOrders: make([]*Order, 0),
		strat:      newStrategy(toIgnore, costMethod),
		prfmLog:    newPrfmLog(outFmt),
		// create channels
		orderChan: make(chan *Order),
		tickChan:  make(chan *Tick),
		closing:   make(chan struct{}, 1),
	}
}

// OMS acts as an `Order Management System` to test trading signals and fill orders.
type OMS struct {
	port       *Portfolio
	benchmark  *Index
	openOrders []*Order
	strat      *strategy
	prfmLog    *PrfmLog
	// Channels
	orderChan chan *Order
	tickChan  chan *Tick
	openChan  chan *Order
	closing   chan struct{}
}

// handle is the function that allows for OMS to integrate as a part of the simulation pipeline.
func (oms *OMS) handle() {
	go func() {
		for oms.orderChan != nil || oms.tickChan != nil || oms.openChan != nil {
			select {
			case tick, ok := <-oms.tickChan:
				if !ok {
					log.Println("OMS's tick channel has been closed.")
					oms.tickChan = nil
					continue
				}
				go oms.processTick(tick)

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
	}()
	<-oms.closing
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
		Price:    oms.port.Active[openOrder.Ticker].positions[0].LastBid.Amount,
		Datetime: oms.port.Active[openOrder.Ticker].positions[0].LastBid.Date,
	}
	go func(order *Order) { oms.orderChan <- order }(newOrder)

	// reduce size of openOrders slice since the 0th element has been closed.
	oms.openOrders = oms.openOrders[0:]
	oms.closeOrders()

	return
}

func (oms *OMS) processTick(tick *Tick) {
	if _, exists := oms.strat.ignore[tick.Ticker]; !exists {
		oms.benchmark.updateSecurity(tick)

		if order, signal := oms.strat.algo.EntryLogic(tick); signal &&
			oms.strat.algo.ValidOrder(oms.port, tick, order) {

			go func() { oms.orderChan <- order }()
		}

		if slice := oms.existsInOrders(tick.Ticker); len(slice) > 0 {

			for _, slicedOrder := range slice {
				go func(openOrder *Order) {
					if order, signal := oms.strat.algo.ExitLogic(tick, openOrder); signal &&
						oms.strat.algo.ValidOrder(oms.port, tick, openOrder) {
						go func() { oms.orderChan <- order }()
					}
				}(slicedOrder)
			}
			oms.port.updatePositions(tick)
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
	canceled
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
