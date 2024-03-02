package strategy

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/types"
	"os"
)

var (
	defaultValidators    = []types.ValidatorStrategy{}
	defaultBlockStrategy = types.BlockStrategy{
		DelayEnable:    false,
		BroadCastDelay: 3000, // 3s
		ModifyEnable:   false,
	}
	defaultAttestStrategy = types.AttestStrategy{
		DelayEnable:    false,
		BroadCastDelay: 3000, // 3s
		ModifyEnable:   false,
	}
)

type Strategy struct {
	Validators []types.ValidatorStrategy `json:"validator"`
	Block      types.BlockStrategy       `json:"block"`
	Attest     types.AttestStrategy      `json:"attest"`
}

func (s *Strategy) GetValidatorRole(valIdx int, slot int64) types.RoleType {
	for _, v := range s.Validators {
		if v.ValidatorIndex == valIdx {
			if slot >= int64(v.AttackerStartSlot) && slot <= int64(v.AttackerEndSlot) {
				return types.AttackerRole
			}
		}
	}
	return types.NormalRole
}

func ParseStrategy(file string) *Strategy {
	var defautConfig = &Strategy{
		Block:  defaultBlockStrategy,
		Attest: defaultAttestStrategy,
	}
	var s Strategy
	d, err := os.ReadFile(file)
	if err != nil {
		log.WithError(err).Error("read strategy failed, use default config")
		return defautConfig
	}
	if err = json.Unmarshal(d, &s); err != nil {
		log.WithError(err).Error("unmarshal strategy failed, use default config")
		return defautConfig
	}
	return &s
}
