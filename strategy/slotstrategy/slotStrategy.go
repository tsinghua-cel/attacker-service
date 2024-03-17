package slotstrategy

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
