package strategy

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/types"
	"os"
)

var (
	defaultValidators    = make([]types.ValidatorStrategy, 0)
	defaultSlots         = make([]types.SlotStrategy, 0)
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

func ParseStrategy(backend types.ServiceBackend, file string) *types.Strategy {
	var defautConfig = &types.Strategy{
		Slots:      defaultSlots,
		Validators: defaultValidators,
		Block:      defaultBlockStrategy,
		Attest:     defaultAttestStrategy,
	}
	var s types.Strategy
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
