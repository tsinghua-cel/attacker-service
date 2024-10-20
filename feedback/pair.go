package feedback

import (
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/strategy/slotstrategy"
	"sync/atomic"
)

type pairStrategy struct {
	uid      string
	parsed   []*slotstrategy.InternalSlotStrategy
	maxEpoch atomic.Value
	minEpoch atomic.Value
}

var (
	FOREVER = int64(1<<63 - 100)
)

// calcEpochs calculate the min and max epoch of the strategy.
func (p *pairStrategy) calcEpochs() (int64, int64) {
	var minEpoch, maxEpoch int64 = FOREVER, -1
	for _, s := range p.parsed {
		switch s.Slot.(type) {
		case *slotstrategy.NumberSlot:
			slot := s.Slot.(*slotstrategy.NumberSlot)
			epoch := common.DefaultSlotTool.SlotToEpoch(int64(*slot))
			if epoch > maxEpoch {
				maxEpoch = epoch
			}
			if epoch < minEpoch {
				minEpoch = epoch
			}
		case *slotstrategy.FunctionSlot:
			maxEpoch = FOREVER

		default:
			// unknown slot type.
		}
	}
	if minEpoch > maxEpoch {
		return maxEpoch, minEpoch
	}
	return minEpoch, maxEpoch
}

func (p *pairStrategy) IsEnd(slot int64) bool {
	var maxEpoch int64
	epoch := common.DefaultSlotTool.SlotToEpoch(slot)
	if v := p.maxEpoch.Load(); v != nil {
		maxEpoch = v.(int64)
	} else {
		// calc max epoch
		mi, ma := p.calcEpochs()
		p.maxEpoch.Store(ma)
		p.minEpoch.Store(mi)
		maxEpoch = ma
	}
	return epoch >= (maxEpoch + 2)
}
