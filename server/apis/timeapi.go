package apis

import (
	"math/rand"
	"time"
)

// TimeAPI offers and API for time operations.
type TimeAPI struct {
	b Backend
}

// NewTimeAPI creates a new tx pool service that gives information about the transaction pool.
func NewTimeAPI(b Backend) *TimeAPI {
	return &TimeAPI{b}
}

// DelayRandom sleep random seconds time and return the time.
func (s *TimeAPI) DelayRandom(min, max int) int {
	var ts = 0
	if min > max {
		ts = 0
	}
	if min == max {
		ts = min
	} else {
		ts = rand.Intn(max-min) + min // 30+50
	}

	time.Sleep(time.Duration(ts) * time.Second)
	return ts
}

// Delay sleep special seconds time and return the time.
func (s *TimeAPI) Delay(ts int) int {
	time.Sleep(time.Duration(ts) * time.Second)
	return ts
}

// Echo will return msg what input.
func (s *TimeAPI) Echo(msg string) string {
	return msg
}
