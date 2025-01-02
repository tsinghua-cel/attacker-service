package ext_exante

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/generator/utils"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

func BlockStrategy(cur, end int, actions map[string]string) {
	point := types.GetPointByName("BlockBeforeBroadCast")
	actions[point] = fmt.Sprintf("%s:%d", "delayWithSecond", (end+1-cur)*common.GetChainBaseInfo().SecondsPerSlot)
}

func GenSlotStrategy(allHacks []interface{}, fullHackDuties []types.ProposerDuty) []types.SlotStrategy {
	fullDuties := make(map[string]bool)
	strategys := make([]types.SlotStrategy, 0)
	for _, subduties := range allHacks {
		duties := subduties.([]types.ProposerDuty)
		//begin, _ := strconv.Atoi(duties[0].Slot)
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
	}
	// add some random strategy.
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
