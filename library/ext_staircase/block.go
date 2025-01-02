package ext_staircase

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/generator/utils"
	"github.com/tsinghua-cel/attacker-service/types"
	"math/rand"
	"strconv"
)

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
			stageII := (rand.Intn(20) + 15) * secondsPerSlot

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

func GenSlotStrategy(hackDuties []types.ProposerDuty, cas int, fullHackDuties []types.ProposerDuty) []types.SlotStrategy {
	fullDuties := make(map[string]bool)
	strategys := make([]types.SlotStrategy, 0)
	for i := 0; i < len(hackDuties); i++ {
		s := getSlotStrategy(hackDuties[i].Slot, cas, i == len(hackDuties)-1)
		strategys = append(strategys, s)
		fullDuties[hackDuties[i].Slot] = true
	}

	for _, duty := range fullHackDuties {
		if _, ok := fullDuties[duty.Slot]; ok {
			continue
		}
		strategy := types.SlotStrategy{
			Slot:    duty.Slot,
			Level:   1,
			Actions: make(map[string]string),
		}
		slot, _ := strconv.Atoi(duty.Slot)
		strategy.Actions = utils.GetRandomActions(slot, utils.SafeRand(4))
		strategys = append(strategys, strategy)
	}
	return strategys
}
