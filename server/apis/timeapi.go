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

// Delay sleep random seconds time and return the time.
func (s *TimeAPI) Delay() int {
	ts := rand.Intn(50) + 30 // 30+50
	time.Sleep(time.Duration(ts) * time.Second)
	return ts
}

// ShortDelay sleep random (3~8)seconds time and return the time.
func (s *TimeAPI) ShortDelay() int {
	ts := rand.Intn(5) + 3 // 3-8
	time.Sleep(time.Duration(ts) * time.Second)
	return ts
}

// Echo will return msg what input.
func (s *TimeAPI) Echo(msg string) string {
	return msg
}
