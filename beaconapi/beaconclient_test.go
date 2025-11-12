package beaconapi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
)

// test GetValidators
func TestGetValidators(t *testing.T) {
	endpoint := "52.221.177.10:34000" // grpc endpoint
	pubks, err := GetValidators(endpoint)
	if err != nil {
		t.Fatalf("get validators failed err:%s", err)
	}
	fmt.Printf("get validators %v\n", pubks)
}

func TestGetGenesisState(t *testing.T) {
	endpoint := "http://18.168.16.120:32946"
	client := NewBeaconGwClient(endpoint)
	state, err := client.GetLatestValidators()
	if err != nil {
		t.Fatalf("get genesis failed err:%v", err)
	}
	vals, err := state.Validators()
	if err != nil {
		t.Fatalf("get validators failed err:%v", err)
	}
	fmt.Printf("get validators %d\n", len(vals))
	for idx, val := range vals {
		fmt.Printf("get validator [%d]:%s\n", idx, val.String())
	}
}

func TestGetAllReward(t *testing.T) {
	endpoint := "52.221.177.10:33500" // grpc gateway endpoint
	client := NewBeaconGwClient(endpoint)
	res, err := client.GetAllValReward(1)
	if err != nil {
		t.Fatalf("get reward failed err:%s", err)
	}
	fmt.Printf("get all reward res:%v\n", res)
}

func TestGetConfig(t *testing.T) {
	endpoint := "52.221.177.10:33500" // grpc gateway endpoint
	client := NewBeaconGwClient(endpoint)
	epoch, err := client.GetIntConfig(SLOTS_PER_EPOCH)
	if err != nil {
		t.Fatalf("get epoch config failed err:%s", err)
	}
	fmt.Printf("get epoch :%d\n", epoch)
}

func TestGetLatestBeaconHeader(t *testing.T) {
	endpoint := "52.221.177.10:33500" // grpc gateway endpoint
	client := NewBeaconGwClient(endpoint)

	header, err := client.GetLatestBeaconHeader()
	if err != nil {
		t.Fatalf("get latest header failed err:%s", err)
	}
	fmt.Printf("get latest header.slot :%s\n", header.Header.Message.Slot)

}

func TestGetAllAttestDuties(t *testing.T) {
	endpoint := "52.221.177.10:14000" // grpc gateway endpoint
	client := NewBeaconGwClient(endpoint)
	duties, err := client.GetProposerDuties(2)
	//duties, err := client.GetCurrentEpochProposerDuties()
	if err != nil {
		t.Fatalf("get proposer duties failed err:%s", err)
	}

	latestSlotWithAttacker := int64(-1)
	for _, duty := range duties {
		dutySlot, _ := strconv.ParseInt(duty.Slot, 10, 64)
		dutyValIdx, _ := strconv.Atoi(duty.ValidatorIndex)
		fmt.Printf("slot=%d, validx =%d\n", dutySlot, dutyValIdx)

		if dutyValIdx <= 31 && dutySlot > latestSlotWithAttacker {
			latestSlotWithAttacker = dutySlot
			fmt.Printf("update latestSlotWithAttacker=%d,\n", dutySlot)
		}
	}

	for _, duty := range duties {
		d, _ := json.Marshal(duty)
		fmt.Printf("get attest duty :%s\n", string(d))
	}
}

func TestGetSignedBlockById(t *testing.T) {
	endpoint := "13.41.176.56:14000" // grpc gateway endpoint
	client := NewBeaconGwClient(endpoint)
	data, err := client.GetDenebBlockBySlot(2)
	if err != nil {
		t.Fatalf("get block failed err:%s", err)
	}
	d, _ := json.MarshalIndent(data, "", "  ")
	fmt.Printf("get block :%s\n", d)
}

func TestGetBlockReward(t *testing.T) {
	endpoint := "13.41.176.56:14000" // grpc gateway endpoint
	client := NewBeaconGwClient(endpoint)
	data, err := client.GetBlockReward(2)
	if err != nil {
		t.Fatalf("get block failed err:%s", err)
	}
	d, _ := json.MarshalIndent(data, "", "  ")
	fmt.Printf("get block :%s\n", d)
}

func TestProposeDuty(t *testing.T) {
	endpoint := "47.242.120.101:34001"
	client := NewBeaconGwClient(endpoint)
	epochs := make([]int, 0)
	//h, err := client.GetLatestBeaconHeader()
	//if err != nil {
	//	t.Fatalf("get latest header failed err:%s", err)
	//}
	//slot, _ := strconv.Atoi(h.Header.Message.Slot)
	//beginEpoch := slot / 32
	//for i := 0; i < 5; i++ {
	//	epochs = append(epochs, beginEpoch+i)
	//}
	epochs = append(epochs, 7)
	epochs = append(epochs, 8)
	epochs = append(epochs, 10)
	epochs = append(epochs, 11)
	epochs = append(epochs, 13)
	epochs = append(epochs, 16)
	epochs = append(epochs, 17)
	epochs = append(epochs, 19)
	epochs = append(epochs, 20)

	//epochs = append(epochs, 5)

	for _, epoch := range epochs {
		duty, err := client.GetEpochProposerDuties(epoch)
		if err != nil {
			t.Errorf("get epoch (%d) proposer duties failed err:%s", epoch, err)
		} else {
			s := "["
			for _, d := range duty {
				s += d.ValidatorIndex + ","
			}
			s += "]"
			fmt.Printf("epoch [%d] duty: %s\n", epoch, s)
			//d, _ := json.Marshal(duty)
			//fmt.Printf("get epoch (%d) proposer duties :%s\n", epoch, string(d))
		}
	}

}

func TestProposeSlotRoot(t *testing.T) {
	endpoint := "47.242.120.101:34003"
	client := NewBeaconGwClient(endpoint)
	root, err := client.GetSlotRoot(64)
	if err != nil {
		t.Fatalf("get slot root failed err:%s", err)
	}
	fmt.Printf("get slot root: %s\n", root)

}

func TestBeaconState(t *testing.T) {
	endpoint := "47.242.120.101:34002"
	client := NewBeaconGwClient(endpoint)
	state, err := client.GetBeaconState("head")
	if err != nil {
		t.Fatalf("get beacon state failed err:%s", err)
	}
	slot, _ := state.Slot()
	fmt.Printf("get beacon state slot: %d\n", slot)
}
