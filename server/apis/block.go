package apis

import (
	"encoding/json"
	"errors"
	"github.com/prysmaticlabs/prysm/v4/cache/lru"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/plugins"
	"github.com/tsinghua-cel/attacker-service/types"
)

var (
	ErrNilObject              = errors.New("nil object")
	ErrUnsupportedBeaconBlock = errors.New("unsupported beacon block")
	blockCacheContent         = lru.New(1000)
)

// BlockAPI offers and API for block operations.
type BlockAPI struct {
	b      Backend
	plugin plugins.AttackerPlugin
}

// NewBlockAPI creates a new tx pool service that gives information about the transaction pool.
func NewBlockAPI(b Backend, plugin plugins.AttackerPlugin) *BlockAPI {
	return &BlockAPI{b, plugin}
}

func (s *BlockAPI) GetStrategy(cliInfo string) []byte {
	d, _ := json.Marshal(s.b.GetStrategy().Block)
	return d
}

func (s *BlockAPI) UpdateStrategy(data []byte) error {
	var blockStrategy types.BlockStrategy
	if err := json.Unmarshal(data, &blockStrategy); err != nil {
		return err
	}
	s.b.GetStrategy().Block = blockStrategy
	log.Infof("block strategy updated to %v\n", blockStrategy)
	return nil
}

func (s *BlockAPI) BroadCastDelay() types.AttackerResponse {
	if s.plugin != nil {
		result := s.plugin.BlockDelayForBroadCast(pluginContext(s.b))
		return types.AttackerResponse{
			Cmd: result.Cmd,
		}
	}
	return types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}
}

func (s *BlockAPI) DelayForReceiveBlock(slot uint64) types.AttackerResponse {
	if s.plugin != nil {
		result := s.plugin.BlockDelayForReceiveBlock(pluginContext(s.b), slot)
		return types.AttackerResponse{
			Cmd: result.Cmd,
		}
	}

	return types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}
}

func (s *BlockAPI) BeforeBroadCast(slot uint64) types.AttackerResponse {
	if s.plugin != nil {
		result := s.plugin.BlockBeforeBroadCast(pluginContext(s.b), slot)
		return types.AttackerResponse{
			Cmd: result.Cmd,
		}
	}

	return types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}
}

func (s *BlockAPI) AfterBroadCast(slot uint64) types.AttackerResponse {
	if s.plugin != nil {
		result := s.plugin.BlockAfterBroadCast(pluginContext(s.b), slot)
		return types.AttackerResponse{
			Cmd: result.Cmd,
		}
	}
	return types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}
}

func (s *BlockAPI) BeforeSign(slot uint64, pubkey string, blockDataBase64 string) types.AttackerResponse {
	s.dumpDuties(slot)
	genericSignedBlock, err := common.Base64ToGenericSignedBlock(blockDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: blockDataBase64,
		}
	}
	if s.plugin != nil {
		result := s.plugin.BlockBeforeSign(pluginContext(s.b), slot, pubkey, genericSignedBlock.GetCapella())
		newBlock, ok := result.Result.(*ethpb.SignedBeaconBlockCapella)
		if ok {
			genericSignedBlock.Block = &ethpb.GenericSignedBeaconBlock_Capella{
				Capella: newBlock,
			}
			newBlockBase64, _ := common.GenericSignedBlockToBase64(genericSignedBlock)
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: newBlockBase64,
			}
		} else {

			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: blockDataBase64,
			}
		}
	}
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: blockDataBase64,
	}
}

func (s *BlockAPI) AfterSign(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	genericSignedBlock, err := common.Base64ToGenericSignedBlock(signedBlockDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: signedBlockDataBase64,
		}
	}
	if s.plugin != nil {
		result := s.plugin.BlockAfterSign(pluginContext(s.b), slot, pubkey, genericSignedBlock.GetCapella())
		newBlock, ok := result.Result.(*ethpb.SignedBeaconBlockCapella)
		if ok {
			genericSignedBlock.Block = &ethpb.GenericSignedBeaconBlock_Capella{
				Capella: newBlock,
			}
			newBlockBase64, _ := common.GenericSignedBlockToBase64(genericSignedBlock)
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: newBlockBase64,
			}
		} else {

			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: signedBlockDataBase64,
			}
		}
	}
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedBlockDataBase64,
	}
}

func (s *BlockAPI) BeforePropose(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	genericSignedBlock, err := common.Base64ToGenericSignedBlock(signedBlockDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: signedBlockDataBase64,
		}
	}
	if s.plugin != nil {
		result := s.plugin.BlockBeforePropose(pluginContext(s.b), slot, pubkey, genericSignedBlock.GetCapella())
		newBlock, ok := result.Result.(*ethpb.SignedBeaconBlockCapella)
		if ok {
			genericSignedBlock.Block = &ethpb.GenericSignedBeaconBlock_Capella{
				Capella: newBlock,
			}
			newBlockBase64, _ := common.GenericSignedBlockToBase64(genericSignedBlock)
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: newBlockBase64,
			}
		} else {

			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: signedBlockDataBase64,
			}
		}
	}
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedBlockDataBase64,
	}
}

func (s *BlockAPI) AfterPropose(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	genericSignedBlock, err := common.Base64ToGenericSignedBlock(signedBlockDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: signedBlockDataBase64,
		}
	}
	if s.plugin != nil {
		result := s.plugin.BlockAfterPropose(pluginContext(s.b), slot, pubkey, genericSignedBlock.GetCapella())
		newBlock, ok := result.Result.(*ethpb.SignedBeaconBlockCapella)
		if ok {
			genericSignedBlock.Block = &ethpb.GenericSignedBeaconBlock_Capella{
				Capella: newBlock,
			}
			newBlockBase64, _ := common.GenericSignedBlockToBase64(genericSignedBlock)
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: newBlockBase64,
			}
		} else {
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: signedBlockDataBase64,
			}
		}
	}
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedBlockDataBase64,
	}
}

func (s *BlockAPI) dumpDuties(slot uint64) {
	slotPerEpoch := s.b.SlotsPerEpoch()
	tool := common.SlotTool{
		SlotsPerEpoch: slotPerEpoch,
	}
	epoch := tool.SlotToEpoch(int(slot))
	if int(slot) == tool.EpochStart(epoch) {
		// dump next epoch duties.
		if slot == 0 {
			if duties, err := s.b.GetProposeDuties(epoch); err == nil {
				for _, duty := range duties {
					log.WithFields(log.Fields{
						"epoch":     epoch,
						"slot":      duty.Slot,
						"validator": duty.ValidatorIndex,
					}).Info("epoch duty")
				}
			}
		}
		if duties, err := s.b.GetProposeDuties(epoch + 1); err == nil {
			for _, duty := range duties {
				log.WithFields(log.Fields{
					"epoch":     epoch,
					"slot":      duty.Slot,
					"validator": duty.ValidatorIndex,
				}).Info("epoch duty")
			}
		}
	}
}
