package livenessnormal

import (
	"github.com/tsinghua-cel/attacker-service/types"
)

func genStrategyForTrigger(epoch int, duties []types.ProposerDuty) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	for _, duty := range duties {
		s := types.SlotStrategy{
			Slot:    duty.Slot,
			Level:   2,
			Actions: make(map[string]string),
		}
		// don't broadcast block.
		s.Actions["BlockBeforeBroadCast"] = "return"
		strategys = append(strategys, s)
	}

	return strategys
}
