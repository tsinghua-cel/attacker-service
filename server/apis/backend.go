package apis

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"github.com/tsinghua-cel/attacker-service/rpc"
	"github.com/tsinghua-cel/attacker-service/strategy"
	types2 "github.com/tsinghua-cel/attacker-service/types"
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
	GetBlockByNumber(number *big.Int) (*types.Block, error)
	GetHeightByNumber(number *big.Int) (*types.Header, error)

	GetValidatorRole(idx int) types2.RoleType
	GetCurrentEpochProposeDuties() ([]beaconapi.ProposerDuty, error)
	GetSlotsPerEpoch() int
	GetIntervalPerSlot() int
}

func GetAPIs(apiBackend Backend) []rpc.API {
	return []rpc.API{
		{
			Namespace: "block",
			Service:   NewBlockAPI(apiBackend),
		},
		{
			Namespace: "attest",
			Service:   NewAttestAPI(apiBackend),
		},
	}
}
