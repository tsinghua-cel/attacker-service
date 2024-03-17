package strategy

import (
	"github.com/tsinghua-cel/attacker-service/types"
)

type internalBlockStrategy struct {
	DelayEnable    bool  `json:"delay_enable"`
	BroadCastDelay int64 `json:"broad_cast_delay"` // unit millisecond
	ModifyEnable   bool  `json:"modify_enable"`
}

func parseToInternalBlockStrategy(strategy types.BlockStrategy) internalBlockStrategy {
	return internalBlockStrategy{
		DelayEnable:    strategy.DelayEnable,
		BroadCastDelay: strategy.BroadCastDelay,
		ModifyEnable:   strategy.ModifyEnable,
	}
}
