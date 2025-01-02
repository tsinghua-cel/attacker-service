package ext_unrealized

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/generator/utils"
	"github.com/tsinghua-cel/attacker-service/types"
	"math/rand"
	"strconv"
)

func GenSlotStrategy(duties []interface{}, fullHackDuties []types.ProposerDuty) []types.SlotStrategy {
	fullDuties := make(map[string]bool)
	strategys := make([]types.SlotStrategy, 0)
	duty := duties[0].([]types.ProposerDuty)
	slotStrategy := types.SlotStrategy{
		Slot:    fmt.Sprintf("%s", duty[0].Slot),
		Level:   1,
		Actions: make(map[string]string),
	}
	currentslot, _ := strconv.Atoi(duty[0].Slot)
	slotStrategy.Actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%d", currentslot-rand.Intn(10)-40)
	strategys = append(strategys, slotStrategy)
	fullDuties[duty[0].Slot] = true
	extendCount := 0
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
		strategy.Actions = utils.GetRandomActions(slot, utils.SafeRand(2))
		strategys = append(strategys, strategy)
		extendCount++
		if extendCount > 2 {
			break
		}
	}

	return strategys

}
