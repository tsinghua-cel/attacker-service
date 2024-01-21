package reward

import (
	"fmt"
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

func TestGetReward(t *testing.T) {
	endpoint := "52.221.177.10:33500" // grpc gateway endpoint
	valIdxs := []int{1, 2, 3, 4, 5}
	client := NewBeaconGwClient(endpoint)
	res, err := client.GetValReward(1, valIdxs)
	if err != nil {
		t.Fatalf("get reward failed err:%s", err)
	}
	fmt.Printf("get specific reward res:%s\n", res)
}

func TestGetAllReward(t *testing.T) {
	endpoint := "52.221.177.10:33500" // grpc gateway endpoint
	client := NewBeaconGwClient(endpoint)
	res, err := client.GetAllValReward(1)
	if err != nil {
		t.Fatalf("get reward failed err:%s", err)
	}
	fmt.Printf("get all reward res:%s\n", res)
}
