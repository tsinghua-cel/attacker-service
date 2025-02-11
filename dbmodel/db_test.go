package dbmodel

import (
	"testing"
)

func init() {
	DbInit("")
}

func TestGetRewardListByValidatorIndex(t *testing.T) {
	list := GetRewardListByValidatorIndex(0)
	t.Log(list)
}
