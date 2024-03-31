package beaconapi

import (
	"fmt"
	"testing"
)

func TestBeaconGwClient_MonitorReorgEvent(t *testing.T) {
	beaconGwClient := NewBeaconGwClient("52.221.177.10:14000")
	ch := beaconGwClient.MonitorReorgEvent()
	for {
		select {
		case reorgEvent := <-ch:
			fmt.Sprintf("reorg event: %v", reorgEvent)
			t.Log(reorgEvent)
		}
	}
}
