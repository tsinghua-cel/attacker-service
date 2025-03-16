package collection

import (
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"testing"
)

func TestGetAttRewards(t *testing.T) {
	// bftest_1
	beacon := "http://18.168.16.120:33401" // prysm beacon node.
	//beacon := "http://18.168.16.120:33406" // lighthouse beacon node.
	//beacon := "http://18.168.16.120:33409" // teku beacon node.
	client := beaconapi.NewBeaconGwClient(beacon)
	slot := 62
	root, err := client.GetSlotRoot(int64(slot))
	if err != nil {
		t.Error("GetSlotRoot failed", "slot", slot, "error", err)
	} else {
		t.Log("GetSlotRoot", "slot", slot, "root", root)
	}
	bre, err := client.GetBlockReward(slot)
	if err != nil {
		t.Error("GetBlockReward failed", "slot", slot, "error", err)
	} else {
		t.Log("GetBlockReward", "slot", slot, "reward", bre)
	}
	//for epoch := 0; epoch < 6; epoch++ {
	//	rewards, err := client.GetAllValReward(int(epoch))
	//	if err != nil {
	//		t.Error("GetAllValReward failed", "epoch", epoch, "error", err)
	//	} else {
	//		t.Log("GetAllValReward", "epoch", epoch, "rewards", rewards)
	//	}
	//}
}

func TestGetBlockRewards(t *testing.T) {
	// bftest_2
	clList := []string{
		"http://18.168.16.120:33432", // node 1 prysm beacon node.
		"http://18.168.16.120:33437", // node 2 prysm beacon node.	client := beaconapi.NewBeaconGwClient(beacon)
		"http://18.168.16.120:33442", // node 2 prysm beacon node.	client := beaconapi.NewBeaconGwClient(beacon)
	}
	for i, beacon := range clList {
		client := beaconapi.NewBeaconGwClient(beacon)
		slot := 7
		t.Log("node index", i)
		root, err := client.GetSlotRoot(int64(slot))
		if err != nil {
			t.Error("GetSlotRoot failed", "slot", slot, "error", err)
		} else {
			t.Log("GetSlotRoot", "slot", slot, "root", root)
		}
		_, err = client.GetBlockReward(slot)
		if err != nil {
			t.Error("GetBlockReward failed", "slot", slot, "error", err)
		}
		//for slot := 10000; slot < 10010; slot++ {
		//	_, err := client.GetBlockReward(slot)

		//}

	}

}
