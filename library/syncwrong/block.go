package syncwrong

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

func getSlotStrategy(epoch int64, slot string, isLatestHackDuty bool) types.SlotStrategy {
	strategy := types.SlotStrategy{
		Slot:    slot,
		Level:   0,
		Actions: make(map[string]string),
	}
	secondPerSlot := common.GetChainBaseInfo().SecondsPerSlot
	slotsPerEpoch := common.GetChainBaseInfo().SlotsPerEpoch
	switch epoch%3 + 1 {
	case 1, 3:
		strategy.Actions["BlockBeforeSign"] = "return"
		strategy.Actions["AttestBeforeSign"] = fmt.Sprintf("return")
		return strategy

	case 2:
		if isLatestHackDuty {
			strategy.Level = 1

			stageI := (slotsPerEpoch/2 + slotsPerEpoch) * secondPerSlot
			//islot, _ := strconv.Atoi(slot)
			//stageI := (slotsPerEpoch - islot%slotsPerEpoch) * secondPerSlot
			//stageII := (32 + 30) * secondPerSlot

			strategy.Actions["AttestBeforeSign"] = fmt.Sprintf("return")

			strategy.Actions["BlockBeforeSign"] = "packPooledAttest"
			strategy.Actions["BlockDelayForReceiveBlock"] = fmt.Sprintf("%s:%d", "delayWithSecond", stageI)
			strategy.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("%s", "delayHalfEpoch")
			//strategy.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("%s:%d", "delayWithSecond", stageII)

		} else {
			strategy.Actions["BlockBeforeSign"] = "return"
			strategy.Actions["AttestAfterSign"] = fmt.Sprintf("addAttestToPool")
			strategy.Actions["AttestBeforePropose"] = fmt.Sprintf("return")
		}
	}
	return strategy
}

func GenSlotStrategy(allDuties []types.ProposerDuty, epoch int64) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	latestDuty := allDuties[len(allDuties)-1]
	laytestDutySlot, _ := strconv.Atoi(latestDuty.Slot)
	epochStart := common.EpochStart(epoch)
	epochEnd := common.EpochEnd(epoch)
	for i := epochStart; i <= epochEnd; i++ {
		s := getSlotStrategy(epoch, strconv.Itoa(int(i)), i == int64(laytestDutySlot))
		strategys = append(strategys, s)
	}
	return strategys
}
