package porttools

func newPerfLog() *PerfLog {
	return &PerfLog{
		closedOrders:    make([]*Order, 0),
		closedPositions: make([]*Position, 0),
		// create channels.
		orderChan: make(chan *Order),
		posChan:   make(chan *Position),
		startQuit: make(chan bool, 1),
	}
}

// PerfLog conducts performance analysis.
type PerfLog struct {
	closedOrders    []*Order
	closedPositions []*Position
	orderChan       chan *Order
	posChan         chan *Position
	startQuit       chan bool
}

func (log *PerfLog) run() {
	done := make(chan struct{})

	go func() {
		for log.orderChan != nil || log.posChan != nil {
			select {
			case order, ok := <-log.orderChan:
				if !ok {
					log.orderChan = nil
					continue
				}
				log.closedOrders = append(log.closedOrders, order)

			case pos, ok := <-log.posChan:
				if !ok {
					log.posChan = nil
					continue
				}
				log.closedPositions = append(log.closedPositions, pos)

			case <-log.startQuit:
				done <- struct{}{}
			}
		}
	}()
	<-done
}

func (log *PerfLog) quit() {
	close(log.orderChan)
	close(log.posChan)
	log.startQuit <- true
}

// - max-drawdown
// - % profitable
// - total num trades
// - winning/losing trades
// - trading period length

// only see different trades -> simulate market that has traders doing different trades and what their aggregate position look like
