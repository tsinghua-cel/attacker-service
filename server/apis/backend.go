package apis

import (
	"github.com/tsinghua-cel/attacker-service/rpc"
)

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	SomeNeedBackend() bool
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
