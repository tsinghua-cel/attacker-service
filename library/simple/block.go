package simple

import (
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"math/rand"
	"strconv"
)

func getSlotStrategy(epoch int64, slot string, isLatestHackDuty bool) types.SlotStrategy {
	strategy := types.SlotStrategy{
		Slot:    slot,
		Level:   0,
		Actions: make(map[string]string),
	}
	iSlot, _ := strconv.Atoi(slot)
	strategy.Actions = GetRandomActions(iSlot, rand.Intn(3))
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
