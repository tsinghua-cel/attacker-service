package common

type SlotTool struct {
	SlotsPerEpoch int
}

func (s SlotTool) SlotToEpoch(slot int) int {
	return slot / s.SlotsPerEpoch
}

func (s SlotTool) EpochEnd(epoch int) int {
	return (epoch+1)*s.SlotsPerEpoch - 1
}

func (s SlotTool) EpochStart(epoch int) int {
	return epoch * s.SlotsPerEpoch
}
