package staircase

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

/*
Eng:
1. The block does not produce blocks in each epoch, and the vote does not broadcast
2. The last evil in each epoch produces blocks, packs all votes, and delays broadcast.
3. Set the block slot to t, receive block: (32 - t mod 32) * 12s, broadcast block: 12 * 12s on the basis of the previous one
*/

func getSlotStrategy(slot string, cas int, isLatestHackSlot bool) types.SlotStrategy {
	strategy := types.SlotStrategy{
		Slot:    slot,
		Level:   0,
		Actions: make(map[string]string),
	}
	base := common.GetChainBaseInfo()
	secondsPerSlot := base.SecondsPerSlot
	slotsPerEpoch := base.SlotsPerEpoch
	switch cas {
	case 0:
		strategy.Actions["BlockBeforeSign"] = "return"
		strategy.Actions["AttestBeforeSign"] = fmt.Sprintf("return")

	case 1:
		if isLatestHackSlot {
			islot, _ := strconv.Atoi(slot)
			stageI := (slotsPerEpoch - islot%slotsPerEpoch) * secondsPerSlot
			stageII := 12 * secondsPerSlot

			strategy.Actions["AttestBeforeSign"] = fmt.Sprintf("return")

			strategy.Actions["BlockBeforeSign"] = "packPooledAttest"
			strategy.Actions["BlockDelayForReceiveBlock"] = fmt.Sprintf("%s:%d", "delayWithSecond", stageI)
			strategy.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("%s:%d", "delayWithSecond", stageII)
		} else {
			strategy.Actions["BlockBeforeSign"] = "return"
			strategy.Actions["AttestAfterSign"] = fmt.Sprintf("addAttestToPool")
			strategy.Actions["AttestBeforeBroadCast"] = fmt.Sprintf("return")
		}
	}
	return strategy

}

func GenSlotStrategy(hackDuties []types.ProposerDuty, cas int) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	for i := 0; i < len(hackDuties); i++ {
		s := getSlotStrategy(hackDuties[i].Slot, cas, i == len(hackDuties)-1)
		strategys = append(strategys, s)
	}
	return strategys
}
