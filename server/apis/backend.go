package apis

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tsinghua-cel/attacker-service/rpc"
	"github.com/tsinghua-cel/attacker-service/strategy"
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
}

func GetAPIs(apiBackend Backend) []rpc.API {
	return []rpc.API{
		{
			Namespace: "time",
			Service:   NewTimeAPI(apiBackend),
		},
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
