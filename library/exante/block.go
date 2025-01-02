package exante

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

func BlockStrategy(cur, end int, actions map[string]string) {
	base := common.GetChainBaseInfo()
	point := types.GetPointByName("BlockBeforeBroadCast")
	actions[point] = fmt.Sprintf("%s:%d", "delayWithSecond", (end+1-cur)*base.SecondsPerSlot)
}

func GenSlotStrategy(allHacks []interface{}) []types.SlotStrategy {
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
		}
	}

	return strategys

}
