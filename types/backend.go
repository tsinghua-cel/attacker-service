package types

import (
	ethtype "github.com/ethereum/go-ethereum/core/types"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"math/big"
)

type ExecuteBackend interface {
	// get data from execute node.
	GetBlockHeight() (uint64, error)
	GetBlockByNumber(number *big.Int) (*ethtype.Block, error)
	GetHeightByNumber(number *big.Int) (*ethtype.Header, error)
}

type BeaconBackend interface {
	GetCurrentEpochProposeDuties() ([]ProposerDuty, error)
	GetSlotsPerEpoch() int
	SlotsPerEpoch() int
	GetIntervalPerSlot() int
	GetValidatorByProposeSlot(slot uint64) (int, error)
	GetProposeDuties(epoch int) ([]ProposerDuty, error)
}

type CacheBackend interface {
	AddSignedAttestation(slot uint64, pubkey string, attestation *ethpb.Attestation)
	AddSignedBlock(slot uint64, pubkey string, block *ethpb.GenericSignedBeaconBlock)
	GetAttestSet(slot uint64) *SlotAttestSet
	GetBlockSet(slot uint64) *SlotBlockSet
	GetValidatorDataSet() *ValidatorDataSet

	GetValidatorRole(slot int, valIdx int) RoleType
	GetValidatorRoleByPubkey(slot int, pubkey string) RoleType
}

type StrategyBackend interface {
	// update strategy
	GetStrategy() *Strategy
	UpdateBlockBroadDelay(milliSecond int64) error
	UpdateAttestBroadDelay(milliSecond int64) error
}

// ServiceBackend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type ServiceBackend interface {
	ExecuteBackend
	BeaconBackend
	CacheBackend
	StrategyBackend
}