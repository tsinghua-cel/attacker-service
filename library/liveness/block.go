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

// 50ms per slot.
func calcDeltaTime(beginEpoch int64, currentSlot int64) int64 {
	return 50 * (currentSlot - common.EpochStart(beginEpoch))
}

func genStrategyForTrigger1(epoch int, attackerDuties []types.ProposerDuty) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	if len(attackerDuties) == 0 {
		return strategys
	}
	releaseSlot := common.EpochEnd(int64(epoch + 2))
	var lastDuty types.ProposerDuty
	for i, duty := range attackerDuties {
		s := types.SlotStrategy{
			Slot:    duty.Slot,
			Level:   2,
			Actions: make(map[string]string),
		}
		if i == 0 {
			// delay first slot for 4 seconds.
			s.Actions["BlockBeforeBroadCast"] = "delayWithSecond:4"
		} else {
			totalDelay := 1000*(common.TimeToSlot(releaseSlot)-common.TimeToSlot(int64(toInt(duty.Slot)))) + calcDeltaTime(int64(epoch), int64(toInt(duty.Slot)))
			stageI := 1000 * 10
			stageII := totalDelay - int64(stageI)
			// modify block parent to last duty.
			s.Actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%s", lastDuty.Slot)
			// set delay for receive block.
			s.Actions["BlockDelayForReceiveBlock"] = fmt.Sprintf("delayWithMilliSecond:%d", stageI)
			// set delay for broadcast block.
			s.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("delayWithMilliSecond:%d", stageII)
			// don't broadcast attest.
			s.Actions["AttestBeforePropose"] = "return"
			// add attest to pool.
			s.Actions["AttestAfterSign"] = fmt.Sprintf("addAttestToPool")
		}
		if i == len(attackerDuties)-1 {
			// pack pooled attestations.
			s.Actions["BlockBeforeSign"] = "packCurrentEpochAttest"
			s.Actions["AttestBeforePropose"] = "null"
		}

		lastDuty = duty
		strategys = append(strategys, s)
	}

	// set last duty to epoch latest duty.
	epochLatestDuty[epoch] = lastDuty
	return fillDefaultStrategy(epoch, strategys)
}

// before genStrategy, need preCompute best maskDuty.
func genStrategyForTrigger2(epoch int, attackerDuties []types.ProposerDuty, maskDuty types.ProposerDuty) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	if len(attackerDuties) == 0 {
		return strategys
	}
	releaseSlot := common.EpochEnd(int64(epoch + 1))
	var lastDuty = epochLatestDuty[epoch-1]

	for i, duty := range attackerDuties {
		s := types.SlotStrategy{
			Slot:    duty.Slot,
			Level:   2,
			Actions: make(map[string]string),
		}
		if i == 0 {
			// delay first slot for 4 seconds.
			s.Actions["BlockBeforeBroadCast"] = "delayWithSecond:4"
		} else if duty.Slot == maskDuty.Slot {
			// don't proposer block.
			s.Actions["BlockBeforeSign"] = "return"
			s.Actions["AttestBeforePropose"] = "return"
		} else {
			totalDelay := 1000*(common.TimeToSlot(releaseSlot)-common.TimeToSlot(int64(toInt(duty.Slot)))) + calcDeltaTime(int64(epoch-1), int64(toInt(duty.Slot)))
			stageI := 1000 * 10
			stageII := totalDelay - int64(stageI)
			// modify block parent to last duty.
			s.Actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%s", lastDuty.Slot)
			// set delay for receive block.
			s.Actions["BlockDelayForReceiveBlock"] = fmt.Sprintf("delayWithMilliSecond:%d", stageI)
			// pack pooled attestations.
			//s.Actions["BlockBeforeSign"] = "packCurrentEpochAttest"
			// set delay for broadcast block.
			s.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("delayWithMilliSecond:%d", stageII)
			s.Actions["AttestBeforePropose"] = "return"

			lastDuty = duty
		}

		strategys = append(strategys, s)
	}

	// set last duty to epoch latest duty.
	epochLatestDuty[epoch] = lastDuty
	return fillDefaultStrategy(epoch, strategys)
}

// before genStrategy, need preCompute best maskDuty.
func genStrategyForTrigger3(epoch int, attackerDuties []types.ProposerDuty, maskDuty types.ProposerDuty) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	if len(attackerDuties) == 0 {
		return strategys
	}
	releaseSlot := common.EpochEnd(int64(epoch))
	var lastDuty = epochLatestDuty[epoch-1]

	for _, duty := range attackerDuties {
		s := types.SlotStrategy{
			Slot:    duty.Slot,
			Level:   2,
			Actions: make(map[string]string),
		}
		if duty.Slot == maskDuty.Slot {
			// don't proposer block.
			s.Actions["BlockBeforeSign"] = "return"
			s.Actions["AttestBeforePropose"] = "return"
		} else {
			totalDelay := 1000*(common.TimeToSlot(releaseSlot)-common.TimeToSlot(int64(toInt(duty.Slot)))) + calcDeltaTime(int64(epoch-2), int64(toInt(duty.Slot)))
			stageI := 1000 * 10
			stageII := totalDelay - int64(stageI)
			// modify block parent to last duty.
			s.Actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%s", lastDuty.Slot)
			// set delay for receive block.
			s.Actions["BlockDelayForReceiveBlock"] = fmt.Sprintf("delayWithMilliSecond:%d", stageI)
			// pack pooled attestations.
			//s.Actions["BlockBeforeSign"] = "packCurrentEpochAttest"
			// set delay for broadcast block.
			s.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("delayWithMilliSecond:%d", stageII)
			s.Actions["AttestBeforePropose"] = "return"

			lastDuty = duty
		}

		strategys = append(strategys, s)
	}

	// set last duty to epoch latest duty.
	epochLatestDuty[epoch] = lastDuty
	return fillDefaultStrategy(epoch, strategys)
}

// before genStrategy, need preCompute best maskDuty.
func generateSimpleStrategy(epoch int, attackerDuties []types.ProposerDuty) []types.SlotStrategy {
	strategys := make([]types.SlotStrategy, 0)
	if len(attackerDuties) == 0 {
		return strategys
	}
	var lastDuty types.ProposerDuty
	for i, duty := range attackerDuties {
		if i == 0 {
			s := types.SlotStrategy{
				Slot:    duty.Slot,
				Level:   2,
				Actions: make(map[string]string),
			}
			// broadcast delay 4s.
			s.Actions["BlockBeforeBroadCast"] = "delayWithSecond:4"
			strategys = append(strategys, s)
		}
		lastDuty = duty
	}

	// set last duty to epoch latest duty.
	epochLatestDuty[epoch] = lastDuty
	return strategys
}

func GenStrategy(triggerring bool, triggeredEpoch int, epoch int, currentDuty []types.ProposerDuty, maskDuty types.ProposerDuty) []types.SlotStrategy {
	if !triggerring {
		return generateSimpleStrategy(epoch, currentDuty)
	}
	offset := (epoch-triggeredEpoch)%3 + 1
	switch offset {
	case 1:
		return genStrategyForTrigger1(epoch, currentDuty)
	case 2:
		return genStrategyForTrigger2(epoch, currentDuty, maskDuty)
	case 3:
		return genStrategyForTrigger3(epoch, currentDuty, maskDuty)
	default:
		// nothing.
		return nil
	}
}
