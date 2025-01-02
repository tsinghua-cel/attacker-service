package randomdelay

import (
	"github.com/tsinghua-cel/attacker-service/generator/utils"
	"github.com/tsinghua-cel/attacker-service/types"
	"math/rand"
	"strconv"
)

func GenSlotStrategy(allHacks []types.ProposerDuty) []types.SlotStrategy {
	if len(allHacks) == 0 {
		return nil
	}
	strategys := make([]types.SlotStrategy, 0)
	for i := 0; i < len(allHacks); i += 2 {
		duty := allHacks[i]
		slot, _ := strconv.ParseInt(duty.Slot, 10, 64)
		strategy := types.SlotStrategy{
			Slot:    duty.Slot,
			Level:   1,
			Actions: make(map[string]string),
		}
		strategy.Actions = utils.GetRandomActions(int(slot), rand.Intn(4)+1)
		strategys = append(strategys, strategy)
	}

	return strategys

}
