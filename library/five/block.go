package five

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/types"
)

/*
Eng:
Look at the block order of the malicious node in the next epoch in advance;
Two malicious nodes are interspersed with a block of honest nodes, so that the parent of the block of the second malicious node points to the slot of the previous malicious node.
delay strategy: the block of the first malicious node, broadcast delay 1 slot;
*/

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
			Slot:    fmt.Sprintf("%d", c.Slot),
			Level:   1,
			Actions: make(map[string]string),
		}
		slotStrategy.Actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%d", a.Slot)
		strategys = append(strategys, slotStrategy)
	}

	return strategys

}
