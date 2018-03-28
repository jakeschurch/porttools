package porttools

import (
	"errors"
	"fmt"
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
		tickChan:  make(chan *Tick, 6000),
		orderChan: make(chan *Order, 1024),
		benchChan: make(chan *Tick, 6000),
		openChan:  make(chan *Order, 1024),
		portChan:  make(chan *Tick, 6000),
		endMux:    make(chan struct{}),
	}
	return &oms
}

// OMS acts as an `Order Management System` to test trading signals and fill orders.
type OMS struct {
	sync.RWMutex
	Port       *Portfolio
	benchmark  *Index
	openOrders []*Order
	strat      strategy
	prfmLog    *PrfmLog
	// Channels
	tickChan  chan *Tick
	orderChan chan *Order
	portChan  chan *Tick
	benchChan chan *Tick
	openChan  chan *Order
	endMux    chan struct{}
}

// mux is the function that allows for OMS to integrate as a part of the simulation pipeline.
func (oms *OMS) mux() {
	// for oms.orderChan != nil || oms.openChan != nil || oms.benchChan != nil || oms.tickChan != nil {
	// 	select {
	// 	case tick, ok := <-oms.tickChan:
	// 		if !ok {
	// 			log.Println("OMS tick channel has been closed.")
	// 			oms.tickChan = nil
	// 			continue
	// 		}
	// 		oms.processTick(tick)

	// 	case tick, ok := <-oms.benchChan:
	// 		if !ok {
	// 			log.Println("OMS benchmark channel has been closed.")
	// 			oms.benchChan = nil
	// 			continue
	// 		}
	// 		// log.Println("updating benchmark...")
	// 		oms.benchmark.updateSecurity(*tick)

	// 	case tick, ok := <-oms.portChan:
	// 		if !ok {
	// 			log.Println("OMS portfolio channel has been closed.")
	// 			oms.portChan = nil
	// 			continue
	// 		}
	// 		// log.Println("updating position...")
	// 		oms.Port.updatePositions(tick)

	// 	case order, ok := <-oms.orderChan:
	// 		if !ok {
	// 			log.Println("OMS order channel has been closed.")
	// 			oms.orderChan = nil
	// 			oms.getResults()
	// 			continue
	// 		}
	// 		oms.Transact(order)

	// 	case openOrder, ok := <-oms.openChan:
	// 		if !ok {
	// 			log.Println("OMS open order channel has been closed.")
	// 			oms.openChan = nil
	// 			continue
	// 		}
	// 		oms.openOrders = append(oms.openOrders, openOrder)

	// 	case <-oms.endMux:
	// 		log.Println("Closing all orders...")
	// 		oms.closeOrders()
	// 		// break

	// 		// default:
	// 		// 	time.Sleep(1 * time.Nanosecond)
	// 	}
	// }
	// log.Println("OMS mux quit")
}

func (oms *OMS) closeHandle() {
	oms.getResults()

	oms.prfmLog.quit()
}

func (oms *OMS) closeOrders(orderChan chan<- *Order) {
	log.Println("Closing all open orders")

	log.Println("Length of openorders: ", len(oms.openOrders))
	for _, openOrder := range oms.openOrders {
		newOrder := &Order{
			Buy:      false,
			Status:   open,
			Logic:    market,
			Ticker:   openOrder.Ticker,
			Volume:   openOrder.Volume,
			Bid:      oms.Port.Active[openOrder.Ticker].positions[0].LastBid.Amount,
			Ask:      oms.Port.Active[openOrder.Ticker].positions[0].LastAsk.Amount,
			Datetime: oms.Port.Active[openOrder.Ticker].positions[0].LastBid.Date,
		}
		oms.Transact(orderChan, newOrder)
	}
	close(oms.orderChan)
}

func writeToOrderCh(orderChan chan<- *Order, order *Order) {
	orderChan <- order
}

func (oms *OMS) checkLogic(done chan<- struct{}, portCh chan<- Tick, orderCh chan<- *Order, benchChan chan<- *Tick, tickChan <-chan *Tick) {
	fmt.Println("Checking logic...")
	ticksChecked := 0
	for range tickChan {
		tick := <-tickChan
		if tick == nil {
			continue
		}

		if _, exists := oms.strat.ignore[tick.Ticker]; !exists {
			oms.benchmark.updateSecurity(*tick)
			if order, signal := oms.strat.algo.EntryLogic(*tick); signal {
				if oms.strat.algo.ValidOrder(oms.Port, order) {
					oms.Transact(orderCh, order)
				}

				if openOrders, orderErr := oms.existsInOrders(tick.Ticker); orderErr == nil {
					// update Portfolio position
					oms.Port.updatePositions(*tick)
					for _, openOrder := range openOrders {
						if order, signal := oms.strat.algo.ExitLogic(*tick, *openOrder); signal {
							if oms.strat.algo.ValidOrder(oms.Port, order) {
								oms.Transact(orderCh, order)
							}
						}
					}
				}
			}
		}
		fmt.Printf("\r%d ticks checked", ticksChecked)
		ticksChecked++
	}
	close(done)
}

// TODO: change channels to queues
func (oms *OMS) processTick(done chan struct{}, tickChan <-chan *Tick) {
	orderChan := make(chan *Order, 10000)
	portChan := make(chan Tick, 10000)
	benchChan := make(chan *Tick, 10000)

	goDone := make(chan struct{})
	go oms.checkLogic(goDone, portChan, orderChan, benchChan, tickChan)

	for {
		select {
		// case tick := <-portChan:
		// 	oms.Port.updatePositions(tick)

		case order := <-orderChan:
			if order.Buy == true {
				oms.Transact(orderChan, order)
			} else {
				oms.openOrders = append(oms.openOrders, order)
			}

		// case benchTick := <-benchChan:
		// 	oms.benchmark.updateSecurity(*benchTick)

		case <-goDone:
			log.Println("CONGRATULATIONS JAKE")
			goto end

			// default:
			// 	time.Sleep(time.Millisecond)
		}
	}
end:
	oms.closeOrders(orderChan)
	oms.getResults()
	log.Println("Im out!")
	close(done)

	// if _, exists := oms.strat.ignore[tick.Ticker]; !exists {
	// 	// Send tick to benchmark chan
	// 	oms.benchChan <- tick

	// 	// Check to see if a buy order can be submitted
	// 	if order, signal := oms.strat.algo.EntryLogic(*tick); signal {
	// 		if oms.strat.algo.ValidOrder(oms.Port, order) {
	// 			oms.orderChan <- order
	// 		}
	// 	}

	// 	if openOrders, orderErr := oms.existsInOrders(tick.Ticker); orderErr == nil {
	// 		// update Portfolio position
	// 		oms.portChan <- tick

	// 		for _, openOrder := range openOrders {
	// 			if order, signal := oms.strat.algo.ExitLogic(*tick, *openOrder); signal {
	// 				if oms.strat.algo.ValidOrder(oms.Port, order) {
	// 					oms.orderChan <- order
	// 				}
	// 			}
	// 		}
	// 	}
	// }
}

func (oms *OMS) existsInOrders(ticker string) ([]*Order, error) {
	orders := make([]*Order, 0)

	for _, order := range oms.openOrders {
		if order.Ticker == ticker {
			orders = append(orders, order)
		}
	}
	if len(orders) == 0 {
		return nil, errors.New("no open orders with ticker name exist")
	}
	return orders, nil
}

// NewMarketOrder returns a new order that will execute at nearest price.
func NewMarketOrder(buy bool, ticker string, bid, ask, volume Amount, datetime time.Time) *Order {
	return &Order{
		Buy:      buy,
		Status:   open,
		Logic:    market,
		Ticker:   ticker,
		Bid:      bid,
		Ask:      ask,
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
	Bid, Ask, Volume Amount
	Datetime         time.Time
}

func (order *Order) toPosition(volume Amount) *Position {
	bid := &datedMetric{Amount: order.Bid, Date: order.Datetime}
	ask := &datedMetric{Amount: order.Ask, Date: order.Datetime}

	return &Position{
		Ticker:   order.Ticker,
		Volume:   volume,
		NumTicks: 1,
		LastBid:  bid, LastAsk: ask,
		AvgBid: bid.Amount, AvgAsk: ask.Amount,
		MaxBid: bid, MaxAsk: ask,
		MinBid: bid, MinAsk: ask,
		BuyPrice: ask,
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
