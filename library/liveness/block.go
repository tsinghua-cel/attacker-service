package liveness

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

var (
	epochLatestDuty = make(map[int]types.ProposerDuty)
)

func fillDefaultStrategy(epoch int, strategies []types.SlotStrategy) []types.SlotStrategy {
	exists := make(map[int]types.SlotStrategy)
	for _, s := range strategies {
		exists[toInt(s.Slot)] = s
	}
	epochStart := common.EpochStart(int64(epoch))
	epochEnd := common.EpochEnd(int64(epoch))
	for i := epochStart; i <= epochEnd; i++ {
		if _, ok := exists[int(i)]; !ok {
			ns := types.SlotStrategy{
				Slot:    strconv.Itoa(int(i)),
				Level:   2,
				Actions: make(map[string]string),
			}
			ns.Actions["AttestBeforePropose"] = "return"
			ns.Actions["AttestAfterSign"] = fmt.Sprintf("addAttestToPool")
			exists[int(i)] = ns
		} else {
			//if _, ok := os.Actions["AttestBeforeBroadCast"]; !ok {
			//	os.Actions["AttestBeforeBroadCast"] = "return"
			//}
		}
	}
	nstrategies := make([]types.SlotStrategy, 0)
	for i := epochStart; i <= epochEnd; i++ {
		if s, ok := exists[int(i)]; ok {
			nstrategies = append(nstrategies, s)
		}
	}
	return nstrategies

}

//// 100ms per slot.
//func calcDeltaTime(beginEpoch int64, currentSlot int64) int64 {
//	return 100 * (currentSlot - common.EpochStart(beginEpoch))
//}

var (
	offsetForSlot = int(0)
	offsetCache   = make(map[int64]map[int64]int) // targetSlot - > (slot -> offset)
)

func calcTargetTime(slot int, targetSlot int64) int64 {
	slotOffset := 0
	if _, ok := offsetCache[targetSlot]; !ok {
		offsetCache[targetSlot] = make(map[int64]int)
		offsetForSlot = 0
		offsetCache[targetSlot][int64(slot)] = offsetForSlot
	} else {
		if _, ok := offsetCache[targetSlot][int64(slot)]; !ok {
			offsetForSlot += 1
			offsetCache[targetSlot][int64(slot)] = offsetForSlot
		}
	}
	slotOffset = offsetCache[targetSlot][int64(slot)]
	begin := common.TimeToSlot(targetSlot+1)*1000 - 8*1000
	return begin - 5*1000 + 100*int64(slotOffset)
}

func genSimpleStrategy(epoch int, attackerDuties []types.ProposerDuty) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	duty := attackerDuties[0]
	s := types.SlotStrategy{
		Slot:    duty.Slot,
		Level:   1,
		Actions: make(map[string]string),
	}
	s.Actions["BlockBeforeBroadCast"] = "delayWithSecond:4"
	s.Actions["AttestBeforePropose"] = "return"
	s.Actions["AttestAfterSign"] = fmt.Sprintf("addAttestToPool")

	strategys = append(strategys, s)
	return strategys
}

func genStrategyForTrigger1(epoch int, attackerDuties []types.ProposerDuty, masked BestMaskDutyInfo) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	if len(attackerDuties) == 0 {
		return strategys
	}
	releaseSlot := common.EpochEnd(int64(epoch + 2))
	var lastDuty types.ProposerDuty
	var maskedDutyMap = make(map[string]bool)
	for _, duty := range masked.maskedDuties {
		maskedDutyMap[duty.Slot] = true
	}

	for i, duty := range attackerDuties {
		s := types.SlotStrategy{
			Slot:    duty.Slot,
			Level:   2,
			Actions: make(map[string]string),
		}
		if i == 0 {
			// delay first slot for 4 seconds.
			s.Actions["BlockBeforeBroadCast"] = "delayWithSecond:4"
			s.Actions["AttestBeforePropose"] = "return"
			s.Actions["AttestAfterSign"] = fmt.Sprintf("addAttestToPool")
		} else if i == len(attackerDuties)-1 {
			// pack pooled attestations.
			s.Actions["BlockBeforeSign"] = "packCurrentEpochAttest"
			s.Actions["AttestBeforePropose"] = "null"
		} else {
			if _, exist := maskedDutyMap[duty.Slot]; exist {
				// don't proposer block.
				s.Actions["BlockBeforeSign"] = "return"
				s.Actions["AttestBeforePropose"] = "return"
			} else {
				targetTime := calcTargetTime(toInt(duty.Slot), releaseSlot)
				// set delay for broadcast block.
				s.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("delayToMilliTime:%d", targetTime)
				// don't broadcast attest.
				s.Actions["AttestBeforePropose"] = "return"
				// add attest to pool.
				s.Actions["AttestAfterSign"] = fmt.Sprintf("addAttestToPool")
			}
		}

		lastDuty = duty
		strategys = append(strategys, s)
	}

	// set last duty to epoch latest duty.
	epochLatestDuty[epoch] = lastDuty
	return fillDefaultStrategy(epoch, strategys)
}

// before genStrategy, need preCompute best maskDuty.
func genStrategyForTrigger2(epoch int, attackerDuties []types.ProposerDuty, masked BestMaskDutyInfo) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	if len(attackerDuties) == 0 {
		return strategys
	}
	releaseSlot := common.EpochEnd(int64(epoch + 1))
	var lastDuty = epochLatestDuty[epoch-1]

	var maskedDutyMap = make(map[string]bool)
	for _, duty := range masked.maskedDuties {
		maskedDutyMap[duty.Slot] = true
	}

	for i, duty := range attackerDuties {
		s := types.SlotStrategy{
			Slot:    duty.Slot,
			Level:   2,
			Actions: make(map[string]string),
		}
		if i == 0 {
			// delay first slot for 4 seconds.
			s.Actions["BlockBeforeBroadCast"] = "delayWithSecond:4"
			s.Actions["AttestBeforePropose"] = "return"
			s.Actions["AttestAfterSign"] = fmt.Sprintf("addAttestToPool")
		} else {
			if _, exist := maskedDutyMap[duty.Slot]; exist {
				// don't proposer block.
				s.Actions["BlockBeforeSign"] = "return"
				s.Actions["AttestBeforePropose"] = "return"
			} else {
				targetTime := calcTargetTime(toInt(duty.Slot), releaseSlot)
				// set delay for broadcast block.
				s.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("delayToMilliTime:%d", targetTime)
				s.Actions["AttestBeforePropose"] = "return"

				lastDuty = duty
			}
		}

		strategys = append(strategys, s)
	}

	// set last duty to epoch latest duty.
	epochLatestDuty[epoch] = lastDuty
	return fillDefaultStrategy(epoch, strategys)
}

// before genStrategy, need preCompute best maskDuty.
func genStrategyForTrigger3(epoch int, attackerDuties []types.ProposerDuty, masked BestMaskDutyInfo) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	if len(attackerDuties) == 0 {
		return strategys
	}
	releaseSlot := common.EpochEnd(int64(epoch))
	var lastDuty = epochLatestDuty[epoch-1]
	var maskedDutyMap = make(map[string]bool)
	for _, duty := range masked.maskedDuties {
		maskedDutyMap[duty.Slot] = true
	}

	for _, duty := range attackerDuties {
		s := types.SlotStrategy{
			Slot:    duty.Slot,
			Level:   2,
			Actions: make(map[string]string),
		}
		if _, exist := maskedDutyMap[duty.Slot]; exist {
			// don't proposer block.
			s.Actions["BlockBeforeSign"] = "return"
			s.Actions["AttestBeforePropose"] = "return"
		} else {
			targetTime := calcTargetTime(toInt(duty.Slot), releaseSlot)
			// set delay for broadcast block.
			s.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("delayToMilliTime:%d", targetTime)
			s.Actions["AttestBeforePropose"] = "return"

			lastDuty = duty
		}

		strategys = append(strategys, s)
	}

	// set last duty to epoch latest duty.
	epochLatestDuty[epoch] = lastDuty
	return fillDefaultStrategy(epoch, strategys)
}
