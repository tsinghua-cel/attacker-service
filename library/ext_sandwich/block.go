package ext_sandwich

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/generator/utils"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

func GenSlotStrategy(duties []interface{}, fullHackDuties []types.ProposerDuty) []types.SlotStrategy {
	fullDuties := make(map[string]bool)
	strategys := make([]types.SlotStrategy, 0)
	for i := 0; i < len(duties); i++ {
		duty := duties[i].([]types.ProposerDuty)
		if len(duty) != 3 {
			continue
		}
		a := duty[0]
		//b := duty[1]
		c := duty[2]

		slotStrategy := types.SlotStrategy{
			Slot:    fmt.Sprintf("%s", c.Slot),
			Level:   1,
			Actions: make(map[string]string),
		}
		slotStrategy.Actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%s", a.Slot)
		strategys = append(strategys, slotStrategy)
		fullDuties[c.Slot] = true
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
