package types

import (
	"encoding/json"
	"os"
)

type ValidatorStrategy struct {
	ValidatorIndex    int `json:"validator_index"`
	AttackerStartSlot int `json:"attacker_start_slot"`
	AttackerEndSlot   int `json:"attacker_end_slot"`
}

type SlotStrategy struct {
	Slot    string            `json:"slot"`
	Level   int               `json:"level"`
	Actions map[string]string `json:"actions"`
}

type Strategy struct {
	Uid        string              `json:"uid"`
	Slots      []SlotStrategy      `json:"slots"`
	Validators []ValidatorStrategy `json:"validator"`
}

func (s Strategy) ToFile(name string) error {
	d, _ := json.MarshalIndent(s, "", "  ")
	return os.WriteFile(name, d, 0644)
}

func (s Strategy) String() string {
	d, _ := json.MarshalIndent(s, "", "  ")
	return string(d)
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

type StrategyGeneratorParam struct {
	Strategy            string `json:"strategy"`
	DurationPerStrategy int64  `json:"duration_per_strategy"`
	MinMaliciousIdx     int    `json:"min_malicious_idx"`
	MaxMaliciousIdx     int    `json:"max_malicious_idx"`
}
