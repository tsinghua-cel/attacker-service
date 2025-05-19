package liveness

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/types"
)

func getSlotStrategy(slot string) types.SlotStrategy {
	strategy := types.SlotStrategy{
		Slot:    slot,
		Level:   1,
		Actions: make(map[string]string),
	}
	stageI := 8
	//strategy.Actions["BlockDelayForReceiveBlock"] = fmt.Sprintf("%s:%d", "delayWithSecond", stageI)
	strategy.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("%s:%d", "delayWithSecond", stageI)
	return strategy
}

func GenSlotStrategy(nextEpochDuties []types.ProposerDuty) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	strategys = append(strategys, getSlotStrategy(nextEpochDuties[0].Slot))
	return strategys
}
