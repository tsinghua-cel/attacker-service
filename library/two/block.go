package two

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/types"
)

func getSlotStrategy(epoch int64, slot string, isLatestHackDuty bool) types.SlotStrategy {
	strategy := types.SlotStrategy{
		Slot:    slot,
		Level:   0,
		Actions: make(map[string]string),
	}
	switch epoch % 3 {
	case 0, 1:
		strategy.Actions["BlockBeforeSign"] = "return"
		strategy.Actions["AttestBeforeSign"] = fmt.Sprintf("return")

	case 2:
		if isLatestHackDuty {

			strategy.Actions["AttestBeforeSign"] = fmt.Sprintf("return")

			strategy.Actions["BlockBeforeSign"] = "packPooledAttest"
			strategy.Actions["BlockDelayForReceiveBlock"] = fmt.Sprintf("delayHalfEpoch")
			// block delay to next epoch-end slot
			strategy.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("delayHalfEpoch")
		} else {
			strategy.Actions["BlockBeforeSign"] = "return"
			strategy.Actions["AttestAfterSign"] = fmt.Sprintf("addAttestToPool")
			strategy.Actions["AttestBeforeBroadCast"] = fmt.Sprintf("return")
		}
	}
	return strategy
}

func GenSlotStrategy(allHackDuties []types.ProposerDuty, epoch int64) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	for i := 0; i < len(allHackDuties); i++ {
		s := getSlotStrategy(epoch, allHackDuties[i].Slot, i == len(allHackDuties)-1)
		strategys = append(strategys, s)
	}
	return strategys

}
