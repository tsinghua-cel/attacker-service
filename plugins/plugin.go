package plugins

import (
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/tsinghua-cel/attacker-service/strategy"
	"github.com/tsinghua-cel/attacker-service/types"
	"github.com/tsinghua-cel/attacker-service/validatorSet"
	"math/big"
)

type BackendContext interface {
// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	SomeNeedBackend() bool
	// update strategy
	GetStrategy() *strategy.Strategy
	UpdateBlockBroadDelay(milliSecond int64) error
	UpdateAttestBroadDelay(milliSecond int64) error

	// get data from execute node.
	GetBlockHeight() (uint64, error)
	GetBlockByNumber(number *big.Int) (*types.Block, error)
	GetHeightByNumber(number *big.Int) (*types.Header, error)

	GetValidatorRole(slot int, valIdx int) types2.RoleType
	GetValidatorRoleByPubkey(slot int, pubkey string) types2.RoleType
	GetCurrentEpochProposeDuties() ([]beaconapi.ProposerDuty, error)
	GetSlotsPerEpoch() int
	SlotsPerEpoch() int
	GetIntervalPerSlot() int
	AddSignedAttestation(slot uint64, pubkey string, attestation *ethpb.Attestation)
	AddSignedBlock(slot uint64, pubkey string, block *ethpb.GenericSignedBeaconBlock)
	GetAttestSet(slot uint64) *validatorSet.SlotAttestSet
	GetBlockSet(slot uint64) *validatorSet.SlotBlockSet
	GetValidatorDataSet() *validatorSet.ValidatorDataSet
	GetValidatorByProposeSlot(slot uint64) (int, error)
	GetProposeDuties(epoch int) ([]beaconapi.ProposerDuty, error)
}

type AttackerPlugin interface {
	AttestBeforeBroadCast(slot uint64) types.AttackerResponse
	AttestAfterBroadCast(slot uint64) types.AttackerResponse
	AttestBeforeSign(slot uint64, pubkey string, attestData *ethpb.AttestationData) types.AttackerResponse
	AttestAfterSign(slot uint64, pubkey string, attest *ethpb.Attestation) types.AttackerResponse
	AttestBeforePropose(slot uint64, pubkey string, attest *ethpb.Attestation) types.AttackerResponse
	AttestAfterPropose(slot uint64, pubkey string, attest *ethpb.Attestation) types.AttackerResponse

	BlockBroadCastDelay() types.AttackerResponse
	BlockDelayForReceiveBlock(slot uint64) types.AttackerResponse
	BlockBeforeBroadCast(slot uint64) types.AttackerResponse
	BlockAfterBroadCast(slot uint64) types.AttackerResponse
	BlockBeforeMakeBlock(slot uint64, pubkey string) types.AttackerResponse
	BlockBeforeSign(slot uint64, pubkey string, block *ethpb.GenericSignedBeaconBlock_Capella) types.AttackerResponse
	BlockAfterSign(slot uint64, pubkey string, block *ethpb.GenericSignedBeaconBlock_Capella) types.AttackerResponse
	BlockBeforePropose(slot uint64, pubkey string, block *ethpb.SignedBeaconBlockCapella) types.AttackerResponse
	BlockAfterPropose(slot uint64, pubkey string, block *ethpb.SignedBeaconBlockCapella) types.AttackerResponse
}
