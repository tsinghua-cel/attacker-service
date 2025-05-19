package common

import (
	"context"
	"sync"
	"time"
)

// TimeClock represents a clock that can notify multiple listeners when a target time is reached
type TimeClock struct {
	mu         sync.Mutex
	targetTime time.Time
	timer      *time.Timer
	listeners  []chan time.Time
	ctx        context.Context
	cancel     context.CancelFunc
}

// New creates a new TimeClock
func NewTimeClock() *TimeClock {
	ctx, cancel := context.WithCancel(context.Background())
	tc := &TimeClock{
		listeners: make([]chan time.Time, 0),
		ctx:       ctx,
		cancel:    cancel,
	}
	go tc.run()
	return tc
}

// AddListener registers a new channel to be notified when the target time is reached
func (tc *TimeClock) AddListener() <-chan time.Time {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	ch := make(chan time.Time, 1)
	tc.listeners = append(tc.listeners, ch)
	return ch
}

// SetTarget sets the target time and resets the internal timer
func (tc *TimeClock) SetTarget(targetTime time.Time) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.targetTime = targetTime
	tc.resetTimerLocked()
}

// ResetTarget resets the target time to now + duration
func (tc *TimeClock) ResetTarget(t time.Duration) {
	d := time.Now().Add(t)
	tc.SetTarget(d)
}

// Stop stops the clock and closes all listener channels
func (tc *TimeClock) Stop() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.timer != nil {
		tc.timer.Stop()
	}
	tc.cancel()

	for _, ch := range tc.listeners {
		close(ch)
	}
	tc.listeners = nil
}

// resetTimerLocked resets the internal timer (must be called with lock held)
func (tc *TimeClock) resetTimerLocked() {
	if tc.timer != nil {
		tc.timer.Stop()
	}

	// Calculate duration until target time
	now := time.Now()
	var d time.Duration
	if tc.targetTime.After(now) {
		d = tc.targetTime.Sub(now)
	} else {
		d = 0 // Fire immediately if target is in the past
	}

	tc.timer = time.AfterFunc(d, tc.timerFired)
}

// timerFired is called when the timer expires
func (tc *TimeClock) timerFired() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	now := time.Now()
	for _, ch := range tc.listeners {
		select {
		case ch <- now:
		default:
			// Skip if channel is blocked
		}
	}
}

// run is the main goroutine managing the clock
func (tc *TimeClock) run() {
	<-tc.ctx.Done()
}
