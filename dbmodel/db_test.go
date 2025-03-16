package dbmodel

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"testing"
)

func init() {
	DbInit("eth:12345678@tcp(127.0.0.1:3306)/eth")
}

func TestAttestReward(t *testing.T) {
	reward := &AttestReward{
		Epoch:          1,
		ValidatorIndex: 1,
		HeadAmount:     1,
		TargetAmount:   1,
		SourceAmount:   1,
	}
	err := NewAttestRewardRepository(orm.NewOrm()).Create(reward)
	if err != nil {
		t.Fatal(err)
	}
	if GetMaxAttestRewardEpoch() != 1 {
		fmt.Println("max epoch is ", GetMaxAttestRewardEpoch())
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
