package livenessnormal

import (
	"github.com/tsinghua-cel/attacker-service/types"
)

func genStrategyForTrigger(epoch int, attackerDuties []types.ProposerDuty) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	if len(attackerDuties) == 0 {
		return strategys
	}
	for _, duty := range attackerDuties {
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
