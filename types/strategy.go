package types

type SlotStrategy struct {
	Slot    string            `json:"slot"`
	Level   int               `json:"level"`
	Actions map[string]string `json:"actions"`
}

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

type Strategy struct {
	Slots      []SlotStrategy      `json:"slots"`
	Validators []ValidatorStrategy `json:"validator"`
	Block      BlockStrategy       `json:"block"`
	Attest     AttestStrategy      `json:"attest"`
}

func (s *Strategy) GetValidatorRole(valIdx int, slot int64) RoleType {
	for _, v := range s.Validators {
		if v.ValidatorIndex == valIdx {
			if slot >= int64(v.AttackerStartSlot) && slot <= int64(v.AttackerEndSlot) {
				return AttackerRole
			}
		}
	}
	return NormalRole
}
