package sandwich

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/types"
)

func GenSlotStrategy(duties []interface{}) []types.SlotStrategy {
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
	}

	return strategys

}
