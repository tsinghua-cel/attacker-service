package one

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

/*
Eng:
Look at the block order of the malicious node in the next epoch in advance;
If there are more than two consecutive malicious nodes, start the strategy;
delay strategy: blockdelay to the next slot after the last malicious node block;
The malicious node's voter starts to do evil, delays the vote, and executes the same strategy as blockdelay.
*/
func BlockStrategy(cur, end int, actions map[string]string) {
	secondPerSlot := common.GetChainBaseInfo().SecondsPerSlot
	point := types.GetPointByName("BlockBeforeBroadCast")
	actions[point] = fmt.Sprintf("%s:%d", "delayWithSecond", (end+1-cur)*secondPerSlot)
}

func AttestStrategy(cur, end int, actions map[string]string) {
	secondPerSlot := common.GetChainBaseInfo().SecondsPerSlot
	point := types.GetPointByName("AttestBeforeBroadCast")
	actions[point] = fmt.Sprintf("%s:%d", "delayWithSecond", (end+1-cur)*secondPerSlot)
}

func GenSlotStrategy(allHacks []interface{}) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	for _, subduties := range allHacks {
		duties := subduties.([]types.ProposerDuty)
		//begin, _ := strconv.Atoi(duties[0].Slot)
		end, _ := strconv.Atoi(duties[len(duties)-1].Slot)

		for i := 0; i < len(duties); i++ {
			slot, _ := strconv.Atoi(duties[i].Slot)
			//idx, _ := strconv.Atoi(duties[i].ValidatorIndex)
			strategy := types.SlotStrategy{
				Slot:    duties[i].Slot,
				Level:   0,
				Actions: make(map[string]string),
			}
			BlockStrategy(slot, end, strategy.Actions)
			AttestStrategy(slot, end, strategy.Actions)
			strategys = append(strategys, strategy)
		}
	}

	return strategys

}
