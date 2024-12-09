package games

import (
	"sync"
	"time"
)

type TradeController struct {
	notify    []func(bool)
	period    time.Duration
	isStarted bool
	mx        sync.Mutex
	stop      chan struct{}
}

func NewTradeController(period time.Duration) *TradeController {
	return &TradeController{
		period: period,
		mx:     sync.Mutex{},
		stop:   make(chan struct{}),
	}
}

func (t *TradeController) SetPeriod(period time.Duration) {
	t.period = period
}

func (t *TradeController) RegisterNotify(f func(bool)) {
	t.notify = append(t.notify, f)
}

func (t *TradeController) StartTradePeriod() {
	for _, fn := range t.notify {
		fn(true)
	}
	t.mx.Lock()
	t.isStarted = true
	t.mx.Unlock()
	select {
	case <-t.stop:
	case <-time.After(t.period):
	}
	t.mx.Lock()
	t.isStarted = false
	t.mx.Unlock()
	for _, fn := range t.notify {
		fn(false)
	}

}

func (t *TradeController) StopTradePeriod() {
	t.mx.Lock()
	if t.isStarted {
		select {
		case t.stop <- struct{}{}:
		case <-time.After(time.Second):
		}
		t.mx.Unlock()
	} else {
		t.mx.Unlock()
		for _, fn := range t.notify {
			fn(false)
		}
	}
}
