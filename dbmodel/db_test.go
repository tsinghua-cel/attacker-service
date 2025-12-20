package dbmodel

import (
	"fmt"
	"testing"
)

func init() {
	DbInit("postgres://eth:12345678@127.0.0.1:5432/eth", "")
}

func TestAttestReward(t *testing.T) {
	reward := &AttestReward{
		Epoch:          1,
		ValidatorIndex: 1,
		HeadAmount:     1,
		TargetAmount:   1,
		SourceAmount:   1,
	}
	err := NewAttestRewardRepository(GetDB()).Create(reward)
	if err != nil {
		t.Fatal(err)
	}
	if GetMaxAttestRewardEpoch(nil) != 1 {
		fmt.Println("max epoch is ", GetMaxAttestRewardEpoch(nil))
		t.Fatal("max epoch error")
	}
	if list := GetRewardListByEpoch(1); len(list) != 1 {
		t.Fatal("get reward list by epoch error")
	}
}

func TestGetRewardListByValidatorIndex(t *testing.T) {
	list := GetRewardListByValidatorIndex(0)
	t.Log(list)
}
