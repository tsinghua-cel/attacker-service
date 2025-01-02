package ext_withholding

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/generator/utils"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

func BlockStrategy(cur, end int, actions map[string]string) {
	base := common.GetChainBaseInfo()
	slotsPerEpoch := base.SlotsPerEpoch
	secondsPerSlot := base.SecondsPerSlot
	endStage := end + 1 - cur
	endStage += slotsPerEpoch * utils.SafeRand(2) / utils.SafeRand(4) // add half epoch.
	point := types.GetPointByName("BlockBeforeBroadCast")
	actions[point] = fmt.Sprintf("%s:%d", "delayWithSecond", endStage*secondsPerSlot)
}

func GenSlotStrategy(allHacks []interface{}, fullHackDuties []types.ProposerDuty) []types.SlotStrategy {
	if len(allHacks) == 0 {
		return nil
	}
	fullDuties := make(map[string]bool)
	strategys := make([]types.SlotStrategy, 0)

	// only use the last subduties
	subduties := allHacks[len(allHacks)-1]

	duties := subduties.([]types.ProposerDuty)
	end, _ := strconv.Atoi(duties[len(duties)-1].Slot)

	for i := 0; i < len(duties); i++ {
		slot, _ := strconv.Atoi(duties[i].Slot)
		strategy := types.SlotStrategy{
			Slot:    duties[i].Slot,
			Level:   1,
			Actions: make(map[string]string),
		}
		BlockStrategy(slot, end, strategy.Actions)
		strategys = append(strategys, strategy)
		fullDuties[duties[i].Slot] = true
	}

	for _, duty := range fullHackDuties {
		if _, ok := fullDuties[duty.Slot]; ok {
			continue
		}
		slot, _ := strconv.Atoi(duty.Slot)
		strategy := types.SlotStrategy{
			Slot:    duty.Slot,
			Level:   1,
			Actions: make(map[string]string),
		}
		strategy.Actions = utils.GetRandomActions(slot, utils.SafeRand(4))
		strategys = append(strategys, strategy)
	}

	return strategys

}
