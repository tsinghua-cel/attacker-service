package generator

import (
	"github.com/tsinghua-cel/attacker-service/types"
)

func GetValidatorStrategy(startIndex, endIndex int, startSlot, endSlot int) []types.ValidatorStrategy {
	res := make([]types.ValidatorStrategy, 0)
	for i := startIndex; i <= endIndex; i++ {
		res = append(res, types.ValidatorStrategy{
			ValidatorIndex:    i,
			AttackerStartSlot: startSlot,
			AttackerEndSlot:   endSlot,
		})
	}
	return res
}
