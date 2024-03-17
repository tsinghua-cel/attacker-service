package slotstrategy

import (
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

type SlotIns interface {
	// if slotIns < slot, return -1
	// if slotIns == slot, return 0
	// if slotIns > slot, return 1
	Compare(slot int64) int
}

type internalSlotStrategy struct {
	Slot    SlotIns           `json:"slot"`
	Actions map[string]string `json:"actions"`
}

func parseToInternalSlotStrategy(backend types.ServiceBackend, strategy []types.SlotStrategy) []internalSlotStrategy {
	is := make([]internalSlotStrategy, len(strategy))
	for i, s := range strategy {
		if n, err := strconv.ParseInt(s.Slot, 10, 64); err == nil {
			is[i].Slot = NumberSlot(n)
		} else {
			calc := GetFunctionSlot(backend, s.Slot)
			is[i].Slot = FunctionSlot{calcFunc: calc}
		}
	}
	return is
}
