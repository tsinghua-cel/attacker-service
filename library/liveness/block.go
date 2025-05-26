package liveness

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
)

var (
	epochLatestDuty = make(map[int]types.ProposerDuty)
)

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
			totalDelay := common.TimeToSlot(releaseSlot) - common.TimeToSlot(int64(toInt(duty.Slot)))
			stageI := 10
			stageII := totalDelay - int64(stageI)
			// modify block parent to last duty.
			s.Actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%s", lastDuty.Slot)
			// set delay for receive block.
			s.Actions["BlockDelayForReceiveBlock"] = fmt.Sprintf("delayWithSecond:%d", stageI)
			// set delay for broadcast block.
			s.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("delayWithSecond:%d", stageII)
			// don't broadcast attest.
			s.Actions["AttestBeforeBroadCast"] = "return"
			// add attest to pool.
			s.Actions["AttestAfterSign"] = fmt.Sprintf("addAttestToPool")
		}
		if i == len(attackerDuties)-1 {
			// pack pooled attestations.
			s.Actions["BlockBeforeSign"] = "packPooledAttest"
			s.Actions["AttestBeforeBroadCast"] = "null"
		}

		lastDuty = duty
		strategys = append(strategys, s)
	}

	// set last duty to epoch latest duty.
	epochLatestDuty[epoch] = lastDuty
	return strategys
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
		} else {
			totalDelay := common.TimeToSlot(releaseSlot) - common.TimeToSlot(int64(toInt(duty.Slot)))
			stageI := 10
			stageII := totalDelay - int64(stageI)
			// modify block parent to last duty.
			s.Actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%s", lastDuty.Slot)
			// set delay for receive block.
			s.Actions["BlockDelayForReceiveBlock"] = fmt.Sprintf("delayWithSecond:%d", stageI)
			// set delay for broadcast block.
			s.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("delayWithSecond:%d", stageII)

			lastDuty = duty
		}

		strategys = append(strategys, s)
	}

	// set last duty to epoch latest duty.
	epochLatestDuty[epoch] = lastDuty
	return strategys
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
		} else {
			totalDelay := common.TimeToSlot(releaseSlot) - common.TimeToSlot(int64(toInt(duty.Slot)))
			stageI := 10
			stageII := totalDelay - int64(stageI)
			// modify block parent to last duty.
			s.Actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%s", lastDuty.Slot)
			// set delay for receive block.
			s.Actions["BlockDelayForReceiveBlock"] = fmt.Sprintf("delayWithSecond:%d", stageI)
			// set delay for broadcast block.
			s.Actions["BlockBeforeBroadCast"] = fmt.Sprintf("delayWithSecond:%d", stageII)

			lastDuty = duty
		}

		strategys = append(strategys, s)
	}

	// set last duty to epoch latest duty.
	epochLatestDuty[epoch] = lastDuty
	return strategys
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
