package common

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeClock(t *testing.T) {
	// Create a new clock
	clock := NewTimeClock()
	defer clock.Stop()

	// Get two listeners
	listener1 := clock.AddListener()

	// Set initial target time (3 seconds from now)
	clock.ResetTarget(time.Now().Add(3 * time.Second))
	var target = time.Now().Add(1 * time.Second)
	times := 0

	for {
		select {
		case t := <-listener1:
			times++
			fmt.Println("Listener 1 notified at:", t)
			clock.ResetTarget(time.Now().Add(time.Second * 5))
			if times == 2 {
				return
			}
		case t := <-time.After(target.Sub(time.Now())):
			fmt.Println("Waiting for notification...at:", t)
			target = time.Now().Add(2 * time.Second)
		}
	}
}
