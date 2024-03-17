package slotstrategy

import (
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

type SlotCalc func(slot int64) int64

type FunctionSlot struct {
	calcFunc SlotCalc
}

func (f FunctionSlot) Compare(slot int64) int {
	cSlot := int64(0)
	if f.calcFunc != nil {
		cSlot = f.calcFunc(slot)
	}
	if cSlot > slot {
		return 1
	}
	if cSlot < slot {
		return -1
	}
	return 0
}

func GetFunctionSlot(backend types.ServiceBackend, name string) SlotCalc {

	switch name {
	case "attackerSlot":
		return func(slot int64) int64 {
			slotsPerEpoch := backend.SlotsPerEpoch()
			tool := common.SlotTool{
				SlotsPerEpoch: slotsPerEpoch,
			}
			epoch := tool.SlotToEpoch(slot)
			duties, err := backend.GetProposeDuties(int(epoch))
			if err != nil {
				return slot + 1
			}

			for _, duty := range duties {
				dutySlot, _ := strconv.ParseInt(duty.Slot, 10, 64)
				dutyValIdx, _ := strconv.Atoi(duty.ValidatorIndex)
				if backend.GetValidatorRole(int(dutySlot), dutyValIdx) == types.AttackerRole && dutySlot == slot {
					return slot
				}
			}
			return slot + 1
		}

	case "lastSlotInCurrentEpoch":
		slotsPerEpoch := backend.SlotsPerEpoch()
		tool := common.SlotTool{
			SlotsPerEpoch: slotsPerEpoch,
		}
		return func(slot int64) int64 {
			epoch := tool.SlotToEpoch(slot)
			return tool.EpochEnd(epoch)
		}
	case "lastSlotInNextEpoch":
		slotsPerEpoch := backend.SlotsPerEpoch()
		tool := common.SlotTool{
			SlotsPerEpoch: slotsPerEpoch,
		}
		return func(slot int64) int64 {
			epoch := tool.SlotToEpoch(slot)
			return tool.EpochEnd(epoch + 1)
		}

	case "firstSlotInCurrentEpoch":
		slotsPerEpoch := backend.SlotsPerEpoch()
		tool := common.SlotTool{
			SlotsPerEpoch: slotsPerEpoch,
		}
		return func(slot int64) int64 {
			epoch := tool.SlotToEpoch(slot)
			return tool.EpochStart(epoch)
		}
	case "firstSlotInNextEpoch":
		slotsPerEpoch := backend.SlotsPerEpoch()
		tool := common.SlotTool{
			SlotsPerEpoch: slotsPerEpoch,
		}
		return func(slot int64) int64 {
			epoch := tool.SlotToEpoch(slot)
			return tool.EpochStart(epoch + 1)
		}
	case "lastAttackerSlotInCurrentEpoch":
		return func(slot int64) int64 {
			slotsPerEpoch := backend.SlotsPerEpoch()
			tool := common.SlotTool{
				SlotsPerEpoch: slotsPerEpoch,
			}
			epoch := tool.SlotToEpoch(slot)
			latestSlotWithAttacker := int64(-1)
			duties, err := backend.GetProposeDuties(int(epoch))
			if err != nil {
				return latestSlotWithAttacker
			}

			for _, duty := range duties {
				dutySlot, _ := strconv.ParseInt(duty.Slot, 10, 64)
				dutyValIdx, _ := strconv.Atoi(duty.ValidatorIndex)
				if backend.GetValidatorRole(int(dutySlot), dutyValIdx) == types.AttackerRole && dutySlot > latestSlotWithAttacker {
					latestSlotWithAttacker = dutySlot
				}
			}
			return latestSlotWithAttacker
		}
	case "lastAttackerSlotInNextEpoch":
		return func(slot int64) int64 {
			slotsPerEpoch := backend.SlotsPerEpoch()
			tool := common.SlotTool{
				SlotsPerEpoch: slotsPerEpoch,
			}
			epoch := tool.SlotToEpoch(slot)
			latestSlotWithAttacker := int64(-1)
			duties, err := backend.GetProposeDuties(int(epoch + 1))
			if err != nil {
				return latestSlotWithAttacker
			}

			for _, duty := range duties {
				dutySlot, _ := strconv.ParseInt(duty.Slot, 10, 64)
				dutyValIdx, _ := strconv.Atoi(duty.ValidatorIndex)
				if backend.GetValidatorRole(int(dutySlot), dutyValIdx) == types.AttackerRole && dutySlot > latestSlotWithAttacker {
					latestSlotWithAttacker = dutySlot
				}
			}
			return latestSlotWithAttacker
		}
	default:
		log.WithField("name", name).Error("unknown function slot name")
		return nil
	}
}
