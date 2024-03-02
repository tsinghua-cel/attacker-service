package apis

import (
	"context"
	ethtype "github.com/ethereum/go-ethereum/core/types"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/plugins"
	"github.com/tsinghua-cel/attacker-service/rpc"
	"github.com/tsinghua-cel/attacker-service/strategy"
	"github.com/tsinghua-cel/attacker-service/types"
	"math/big"
)

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
	GetBlockByNumber(number *big.Int) (*ethtype.Block, error)
	GetHeightByNumber(number *big.Int) (*ethtype.Header, error)

	GetValidatorRole(slot int, valIdx int) types.RoleType
	GetValidatorRoleByPubkey(slot int, pubkey string) types.RoleType
	GetCurrentEpochProposeDuties() ([]types.ProposerDuty, error)
	GetSlotsPerEpoch() int
	SlotsPerEpoch() int
	GetIntervalPerSlot() int
	AddSignedAttestation(slot uint64, pubkey string, attestation *ethpb.Attestation)
	AddSignedBlock(slot uint64, pubkey string, block *ethpb.GenericSignedBeaconBlock)
	GetAttestSet(slot uint64) *types.SlotAttestSet
	GetBlockSet(slot uint64) *types.SlotBlockSet
	GetValidatorDataSet() *types.ValidatorDataSet
	GetValidatorByProposeSlot(slot uint64) (int, error)
	GetProposeDuties(epoch int) ([]types.ProposerDuty, error)
}

func GetAPIs(apiBackend Backend, plugin plugins.AttackerPlugin) []rpc.API {
	return []rpc.API{
		{
			Namespace: "admin",
			Service:   NewAdminAPI(apiBackend, plugin),
		},
		{
			Namespace: "block",
			Service:   NewBlockAPI(apiBackend, plugin),
		},
		{
			Namespace: "attest",
			Service:   NewAttestAPI(apiBackend, plugin),
		},
	}
}

func pluginContext(backend types.ServiceBackend) plugins.PluginContext {
	return plugins.PluginContext{
		Backend: backend,
		Context: context.Background(),
		Logger:  log.WithField("module", "attacker-service"),
	}
}
