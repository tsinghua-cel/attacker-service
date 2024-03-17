package strategy

import "github.com/tsinghua-cel/attacker-service/types"

type internalAttestStrategy struct {
	DelayEnable    bool  `json:"delay_enable"`
	BroadCastDelay int64 `json:"broad_cast_delay"` // unit millisecond
	ModifyEnable   bool  `json:"modify_enable"`
}

func parseToInternalAttestStrategy(strategy types.AttestStrategy) internalAttestStrategy {
	return internalAttestStrategy{
		DelayEnable:    strategy.DelayEnable,
		BroadCastDelay: strategy.BroadCastDelay,
		ModifyEnable:   strategy.ModifyEnable,
	}
}
