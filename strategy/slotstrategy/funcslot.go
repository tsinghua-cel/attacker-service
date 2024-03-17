package slotstrategy

import (
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/plugins"
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

func getFunctionSlot(backend plugins.PluginContext, name string) SlotCalc {
	slotsPerEpoch := backend.Backend.SlotsPerEpoch()
	switch name {
	case "lastSlotInCurrentEpoch":
		return func(slot int64) int64 {
			tool := common.SlotTool{
				SlotsPerEpoch: slotsPerEpoch,
			}
			epoch := tool.SlotToEpoch(slot)
			return tool.EpochEnd(epoch)
		}
	case "lastSlotInNextEpoch":
		return func(slot int64) int64 {
			tool := common.SlotTool{
				SlotsPerEpoch: slotsPerEpoch,
			}
			epoch := tool.SlotToEpoch(slot)
			return tool.EpochEnd(epoch + 1)
		}

	case "firstSlotInCurrentEpoch":
		return func(slot int64) int64 {
			tool := common.SlotTool{
				SlotsPerEpoch: slotsPerEpoch,
			}
			epoch := tool.SlotToEpoch(slot)
			return tool.EpochStart(epoch)
		}
	case "firstSlotInNextEpoch":
		return func(slot int64) int64 {
			tool := common.SlotTool{
				SlotsPerEpoch: slotsPerEpoch,
			}
			epoch := tool.SlotToEpoch(slot)
			return tool.EpochStart(epoch + 1)
		}
	case "lastAttackerSlotInCurrentEpoch":
		return func(slot int64) int64 {
			backend.Backend.
		}
	}

}

func isTheLastOneInCurrentEpoch() {

}
