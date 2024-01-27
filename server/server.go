package server

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"github.com/tsinghua-cel/attacker-service/config"
	"github.com/tsinghua-cel/attacker-service/rpc"
	"github.com/tsinghua-cel/attacker-service/server/apis"
	"github.com/tsinghua-cel/attacker-service/strategy"
	types2 "github.com/tsinghua-cel/attacker-service/types"
	"github.com/tsinghua-cel/attacker-service/validatorSet"
	"math/big"
	"strconv"
	"time"
)

type Server struct {
	config       *config.Config
	rpcAPIs      []rpc.API   // List of APIs currently provided by the node
	http         *httpServer //
	strategy     *strategy.Strategy
	execClient   *ethclient.Client
	beaconClient *beaconapi.BeaconGwClient

	validatorSetInfo *validatorSet.ValidatorSet
}

func NewServer() *Server {
	s := &Server{}
	s.config = config.GetConfig()
	s.rpcAPIs = apis.GetAPIs(s)
	client, err := ethclient.Dial(s.config.ExecuteRpc)
	if err != nil {
		panic(fmt.Sprintf("dial execute failed with err:%v", err))
	}
	s.execClient = client
	s.beaconClient = beaconapi.NewBeaconGwClient(s.config.BeaconRpc)
	s.http = newHTTPServer(log.WithField("module", "server"), rpc.DefaultHTTPTimeouts)
	s.strategy = strategy.ParseStrategy(config.GetConfig().Strategy)
	s.validatorSetInfo = validatorSet.NewValidatorSet()
	return s
}

// startRPC is a helper method to configure all the various RPC endpoints during node
// startup. It's not meant to be called at any time afterwards as it makes certain
// assumptions about the state of the node.
func (n *Server) startRPC() error {
	// Filter out personal api
	var (
		servers []*httpServer
	)

	rpcConfig := rpcEndpointConfig{
		batchItemLimit:         config.APIBatchItemLimit,
		batchResponseSizeLimit: config.APIBatchResponseSizeLimit,
	}

	initHttp := func(server *httpServer, port int) error {
		if err := server.setListenAddr(n.config.HttpHost, port); err != nil {
			return err
		}
		if err := server.enableRPC(n.rpcAPIs, httpConfig{
			CorsAllowedOrigins: config.DefaultCors,
			Vhosts:             config.DefaultVhosts,
			Modules:            config.DefaultModules,
			prefix:             config.DefaultPrefix,
			rpcEndpointConfig:  rpcConfig,
		}); err != nil {
			return err
		}
		servers = append(servers, server)
		return nil
	}

	// Set up HTTP.
	// Configure legacy unauthenticated HTTP.
	if err := initHttp(n.http, n.config.HttpPort); err != nil {
		return err
	}

	// Start the servers
	for _, server := range servers {
		if err := server.start(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) monitorDuties() {
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			duties, err := s.beaconClient.GetCurrentEpochAttestDuties()
			if err != nil {
				continue
			}
			for _, duty := range duties {
				idx, _ := strconv.Atoi(duty.ValidatorIndex)
				s.validatorSetInfo.AddValidator(idx, duty.Pubkey, types2.NormalRole)
			}

			ticker.Reset(time.Second * 2)

		}
	}
}

func (s *Server) Start() {
	// start RPC endpoints
	err := s.startRPC()
	if err != nil {
		s.stopRPC()
	}
	// start collect duties info.
	go s.monitorDuties()
}

func (s *Server) stopRPC() {
	s.http.stop()
}

// implement backend
func (s *Server) SomeNeedBackend() bool {
	return true
}

func (s *Server) GetBlockHeight() (uint64, error) {
	return s.execClient.BlockNumber(context.Background())
}

func (s *Server) GetBlockByNumber(number *big.Int) (*types.Block, error) {
	return s.execClient.BlockByNumber(context.Background(), number)
}

func (s *Server) GetHeightByNumber(number *big.Int) (*types.Header, error) {
	return s.execClient.HeaderByNumber(context.Background(), number)
}

func (s *Server) GetStrategy() *strategy.Strategy {
	return s.strategy
}

func (s *Server) UpdateBlockBroadDelay(milliSecond int64) error {
	s.strategy.Block.BroadCastDelay = milliSecond
	return nil
}

func (s *Server) UpdateAttestBroadDelay(milliSecond int64) error {
	s.strategy.Attest.BroadCastDelay = milliSecond
	return nil
}

func (s *Server) GetValidatorRole(idx int) types2.RoleType {
	if val := s.validatorSetInfo.GetValidatorByIndex(idx); val != nil {
		return val.Role
	} else {
		return types2.NormalRole
	}
}

func (s *Server) GetCurrentEpochProposeDuties() ([]beaconapi.ProposerDuty, error) {
	return s.beaconClient.GetCurrentEpochProposerDuties()
}

func (s *Server) GetCurrentEpochAttestDuties() ([]beaconapi.AttestDuty, error) {
	return s.beaconClient.GetCurrentEpochAttestDuties()
}

func (s *Server) GetSlotsPerEpoch() int {
	count, err := s.beaconClient.GetIntConfig(beaconapi.SLOTS_PER_EPOCH)
	if err != nil {
		return 6
	}
	return count
}

func (s *Server) GetIntervalPerSlot() int {
	interval, err := s.beaconClient.GetIntConfig(beaconapi.SECONDS_PER_SLOT)
	if err != nil {
		return 12
	}
	return interval
}
