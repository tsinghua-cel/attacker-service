package validatorSet

import (
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/tsinghua-cel/attacker-service/types"
	"sync"
)

type ValidatorInfo struct {
	Index   int64              `json:"index"`
	Pubkey  string             `json:"pubkey"`
	Role    types.RoleType     `json:"role"`
	Attests ValidatorAttestSet `json:"attests"`
	Blocks  ValidatorBlockSet  `json:"blocks"`
}

type ValidatorSet struct {
	ValidatorByIndex  sync.Map //map[int]*ValidatorInfo
	ValidatorByPubkey sync.Map //map[string]*ValidatorInfo
}

func NewValidatorSet() *ValidatorSet {
	return &ValidatorSet{}
}

func (vs *ValidatorSet) AddValidator(index int, pubkey string, role types.RoleType) {
	v := &ValidatorInfo{
		Index:  int64(index),
		Pubkey: pubkey,
		Role:   role,
	}
	vs.ValidatorByIndex.Store(index, v)
	vs.ValidatorByPubkey.Store(pubkey, v)
}

func (vs *ValidatorSet) GetValidatorByIndex(index int) *ValidatorInfo {
	if v, exist := vs.ValidatorByIndex.Load(index); !exist {
		return nil
	} else {
		return v.(*ValidatorInfo)
	}
}

func (vs *ValidatorSet) GetValidatorByPubkey(pubkey string) *ValidatorInfo {
	if v, exist := vs.ValidatorByPubkey.Load(pubkey); !exist {
		return nil
	} else {
		return v.(*ValidatorInfo)
	}
}

type ValidatorAttestSet struct {
	Attestations map[int]*ethpb.Attestation
}

type ValidatorBlockSet struct {
	Blocks map[int]ethpb.GenericBeaconBlock
}
