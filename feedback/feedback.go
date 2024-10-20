package feedback

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/tsinghua-cel/attacker-service/strategy/slotstrategy"
	"github.com/tsinghua-cel/attacker-service/types"
	"sync"
	"time"
)

type Feedback struct {
	historyStrategy map[int64]*pairStrategy
	quit            chan struct{}
	mux             sync.Mutex
	feed            event.Feed
	backend         types.CacheBackend
}

func NewFeedback(backend types.CacheBackend) *Feedback {
	return &Feedback{
		historyStrategy: make(map[int64]*pairStrategy),
		quit:            make(chan struct{}),
		backend:         backend,
	}
}

func (f *Feedback) AddNewStrategy(uid string, parsed []*slotstrategy.InternalSlotStrategy) {
	f.mux.Lock()
	defer f.mux.Unlock()

	f.historyStrategy[time.Now().UnixMilli()] = &pairStrategy{uid: uid, parsed: parsed}
}

func (f *Feedback) Start() {
	go f.loop()
}

func (f *Feedback) Stop() {
	close(f.quit)
}

type StrategyEndEvent struct {
	Uid      string
	MinEpoch int64
	MaxEpoch int64
}

func (f *Feedback) SubscribeStrategyEndEvent(ch chan StrategyEndEvent) event.Subscription {
	return f.feed.Subscribe(ch)
}

func (f *Feedback) loop() {
	tc := time.NewTicker(time.Second * 10)
	defer tc.Stop()
	for {
		select {
		case <-tc.C:
			f.mux.Lock()
			unrechedTime := int64(0)
			for timestamp, pair := range f.historyStrategy {
				if unrechedTime != 0 && timestamp > unrechedTime {
					continue
				}
				curSlot := f.backend.GetCurSlot()
				if pair.IsEnd(curSlot) {
					// send event
					f.feed.Send(StrategyEndEvent{
						Uid:      pair.uid,
						MinEpoch: pair.minEpoch.Load().(int64),
						MaxEpoch: pair.maxEpoch.Load().(int64),
					})
					delete(f.historyStrategy, timestamp)
					if timestamp < unrechedTime {
						unrechedTime = timestamp
					}
				}
			}
			f.mux.Unlock()
		case <-f.quit:
			return
		}

	}

}
