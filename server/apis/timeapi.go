package apis

import (
	"github.com/tsinghua-cel/attacker-service/types"
	"math/rand"
	"strconv"
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
func (s *TimeAPI) DelayRandom(min, max int) types.AttackerResponse {
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
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: strconv.Itoa(ts),
	}
}

// Delay sleep special seconds time and return the time.
func (s *TimeAPI) Delay(ts int) types.AttackerResponse {
	time.Sleep(time.Duration(ts) * time.Second)
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: strconv.Itoa(ts),
	}
}

// Echo will return msg what input.
func (s *TimeAPI) Echo(msg string) string {
	return msg
}

//
//// Echo will return msg what input.
//func (s *TimeAPI) Echo(msg string) types.AttackerResponse {
//	return types.AttackerResponse{
//		Cmd:    types.CMD_NULL,
//		Result: msg,
//	}
//}
