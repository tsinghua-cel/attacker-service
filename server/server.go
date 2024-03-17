package server

import (
	"context"
	"errors"
	"fmt"
	ethtype "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"github.com/tsinghua-cel/attacker-service/config"
	"github.com/tsinghua-cel/attacker-service/plugins"
	"github.com/tsinghua-cel/attacker-service/rpc"
	"github.com/tsinghua-cel/attacker-service/server/apis"
	"github.com/tsinghua-cel/attacker-service/strategy"
	"github.com/tsinghua-cel/attacker-service/strategy/slotstrategy"
	"github.com/tsinghua-cel/attacker-service/types"
	"math/big"
	"strconv"
	"time"
)

type Server struct {
	config       *config.Config
	rpcAPIs      []rpc.API   // List of APIs currently provided by the node
	http         *httpServer //
	strategy     *types.Strategy
	internal     []slotstrategy.InternalSlotStrategy
	execClient   *ethclient.Client
	beaconClient *beaconapi.BeaconGwClient

	validatorSetInfo *types.ValidatorDataSet
}

func NewServer(conf *config.Config, plugin plugins.AttackerPlugin) *Server {
	s := &Server{}
	s.config = conf
	s.rpcAPIs = apis.GetAPIs(s, plugin)
	client, err := ethclient.Dial(conf.ExecuteRpc)
	if err != nil {
		panic(fmt.Sprintf("dial execute failed with err:%v", err))
	}
	s.execClient = client
	s.beaconClient = beaconapi.NewBeaconGwClient(conf.BeaconRpc)
	s.http = newHTTPServer(log.WithField("module", "server"), rpc.DefaultHTTPTimeouts)
	s.strategy = strategy.ParseStrategy(s, conf.Strategy)
	s.validatorSetInfo = types.NewValidatorSet()
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
				s.validatorSetInfo.AddValidator(idx, duty.Pubkey)
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

func (s *Server) GetBlockByNumber(number *big.Int) (*ethtype.Block, error) {
	return s.execClient.BlockByNumber(context.Background(), number)
}

func (s *Server) GetHeightByNumber(number *big.Int) (*ethtype.Header, error) {
	return s.execClient.HeaderByNumber(context.Background(), number)
}

func (s *Server) GetStrategy() *types.Strategy {
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

func (s *Server) GetValidatorRoleByPubkey(slot int, pubkey string) types.RoleType {
	if val := s.validatorSetInfo.GetValidatorByPubkey(pubkey); val != nil {
		return s.GetValidatorRole(slot, int(val.Index))
	} else {
		return types.NormalRole
	}
}

func (s *Server) GetCurrentEpochProposeDuties() ([]types.ProposerDuty, error) {
	return s.beaconClient.GetCurrentEpochProposerDuties()
}

func (s *Server) GetCurrentEpochAttestDuties() ([]types.AttestDuty, error) {
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

func (s *Server) AddSignedAttestation(slot uint64, pubkey string, attestation *ethpb.Attestation) {
	s.validatorSetInfo.AddSignedAttestation(slot, pubkey, attestation)
}

func (s *Server) AddSignedBlock(slot uint64, pubkey string, block *ethpb.GenericSignedBeaconBlock) {
	s.validatorSetInfo.AddSignedBlock(slot, pubkey, block)
}

func (s *Server) GetAttestSet(slot uint64) *types.SlotAttestSet {
	return s.validatorSetInfo.GetAttestSet(slot)
}

func (s *Server) GetBlockSet(slot uint64) *types.SlotBlockSet {
	return s.validatorSetInfo.GetBlockSet(slot)
}

func (s *Server) GetValidatorDataSet() *types.ValidatorDataSet {
	return s.validatorSetInfo
}

func (s *Server) GetValidatorByProposeSlot(slot uint64) (int, error) {
	epochPerSlot := uint64(s.GetSlotsPerEpoch())
	epoch := slot / epochPerSlot
	duties, err := s.beaconClient.GetProposerDuties(int(epoch))
	if err != nil {
		return 0, err
	}
	for _, duty := range duties {
		dutySlot, _ := strconv.ParseInt(duty.Slot, 10, 64)
		if uint64(dutySlot) == slot {
			idx, _ := strconv.Atoi(duty.ValidatorIndex)
			return idx, nil
		}
	}
	return 0, errors.New("not found")
}

func (s *Server) GetProposeDuties(epoch int) ([]types.ProposerDuty, error) {
	return s.beaconClient.GetProposerDuties(epoch)
}

func (s *Server) SlotsPerEpoch() int {
	return s.GetSlotsPerEpoch()
}

func (s *Server) GetValidatorRole(slot int, valIdx int) types.RoleType {
	if slot < 0 {
		header, err := s.beaconClient.GetLatestBeaconHeader()
		if err != nil {
			return types.NormalRole
		}
		slot, _ = strconv.Atoi(header.Header.Message.Slot)
	}
	return s.strategy.GetValidatorRole(valIdx, int64(slot))
}

func (s *Server) GetInternalSlotStrategy() []slotstrategy.InternalSlotStrategy {
	if len(s.internal) == 0 {
		s.internal = slotstrategy.ParseToInternalSlotStrategy(s, s.strategy.Slots)
	}
	return s.internal
}
