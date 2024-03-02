package types

type ValidatorStrategy struct {
	ValidatorIndex    int `json:"validator_index"`
	AttackerStartSlot int `json:"attacker_start_slot"`
	AttackerEndSlot   int `json:"attacker_end_slot"`
}

type BlockStrategy struct {
	DelayEnable    bool  `json:"delay_enable"`
	BroadCastDelay int64 `json:"broad_cast_delay"` // unit millisecond
	ModifyEnable   bool  `json:"modify_enable"`
}

type AttestStrategy struct {
	DelayEnable    bool  `json:"delay_enable"`
	BroadCastDelay int64 `json:"broad_cast_delay"` // unit millisecond
	ModifyEnable   bool  `json:"modify_enable"`
}
