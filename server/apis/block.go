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

func (s *BlockAPI) BroadCastDelay(slot uint64) types.AttackerResponse {
	result := types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}

	if t, find := findMaxLevelStrategy(s.b.GetInternalSlotStrategy(), int64(slot)); find {
		action := t.Actions["BlockDelayForBroadCast"]
		if action != nil {
			r := action.RunAction(s.b, int64(slot), "")
			result.Cmd = r.Cmd
		}
	}

	return result
}

func (s *BlockAPI) DelayForReceiveBlock(slot uint64) types.AttackerResponse {
	result := types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}

	if t, find := findMaxLevelStrategy(s.b.GetInternalSlotStrategy(), int64(slot)); find {
		action := t.Actions["BlockDelayForReceiveBlock"]
		if action != nil {
			log.WithFields(log.Fields{
				"slot":   slot,
				"action": "BlockDelayForReceiveBlock",
			}).Info("do action")
			r := action.RunAction(s.b, int64(slot), "")
			result.Cmd = r.Cmd
		}
	}

	return result
}

func (s *BlockAPI) BeforeBroadCast(slot uint64) types.AttackerResponse {
	result := types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}

	if t, find := findMaxLevelStrategy(s.b.GetInternalSlotStrategy(), int64(slot)); find {
		action := t.Actions["BlockBeforeBroadCast"]
		if action != nil {
			log.WithFields(log.Fields{
				"slot":   slot,
				"action": "BlockBeforeBroadCast",
			}).Info("do action")
			r := action.RunAction(s.b, int64(slot), "")
			result.Cmd = r.Cmd
		}
	}

	return result
}

func (s *BlockAPI) AfterBroadCast(slot uint64) types.AttackerResponse {
	result := types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}

	if t, find := findMaxLevelStrategy(s.b.GetInternalSlotStrategy(), int64(slot)); find {
		action := t.Actions["BlockAfterBroadCast"]
		if action != nil {
			r := action.RunAction(s.b, int64(slot), "")
			result.Cmd = r.Cmd
		}
	}

	return result
}

func (s *BlockAPI) BeforeSign(slot uint64, pubkey string, blockDataBase64 string) types.AttackerResponse {
	genericSignedBlock, err := common.Base64ToGenericSignedBlock(blockDataBase64)
	if err != nil {
		log.WithError(err).WithField("slot", slot).Error("unmarshal block failed in BeforeSign")
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: blockDataBase64,
		}
	}
	result := types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: blockDataBase64,
	}

	if t, find := findMaxLevelStrategy(s.b.GetInternalSlotStrategy(), int64(slot)); find {
		action := t.Actions["BlockBeforeSign"]
		if action != nil {
			r := action.RunAction(s.b, int64(slot), pubkey, genericSignedBlock.GetCapella())
			result.Cmd = r.Cmd
			newBlock, ok := r.Result.(*ethpb.SignedBeaconBlockCapella)
			if ok {
				genericSignedBlock.Block = &ethpb.GenericSignedBeaconBlock_Capella{
					Capella: newBlock,
				}
				newBlockBase64, _ := common.GenericSignedBlockToBase64(genericSignedBlock)
				result.Result = newBlockBase64
			}
		}
	}

	return result
}

func (s *BlockAPI) AfterSign(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	genericSignedBlock, err := common.Base64ToGenericSignedBlock(signedBlockDataBase64)
	if err != nil {
		log.WithError(err).WithField("slot", slot).Error("unmarshal block failed in after sign")
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: signedBlockDataBase64,
		}
	}
	result := types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedBlockDataBase64,
	}

	if t, find := findMaxLevelStrategy(s.b.GetInternalSlotStrategy(), int64(slot)); find {
		action := t.Actions["BlockAfterSign"]
		if action != nil {
			r := action.RunAction(s.b, int64(slot), pubkey, genericSignedBlock.GetCapella())
			result.Cmd = r.Cmd
			newBlock, ok := r.Result.(*ethpb.SignedBeaconBlockCapella)
			if ok {
				genericSignedBlock.Block = &ethpb.GenericSignedBeaconBlock_Capella{
					Capella: newBlock,
				}
				newBlockBase64, _ := common.GenericSignedBlockToBase64(genericSignedBlock)
				result.Result = newBlockBase64
			}
		}
	}

	return result
}

func (s *BlockAPI) BeforePropose(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	genericSignedBlock, err := common.Base64ToGenericSignedBlock(signedBlockDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: signedBlockDataBase64,
		}
	}
	result := types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedBlockDataBase64,
	}

	if t, find := findMaxLevelStrategy(s.b.GetInternalSlotStrategy(), int64(slot)); find {
		action := t.Actions["BlockBeforePropose"]
		if action != nil {
			r := action.RunAction(s.b, int64(slot), pubkey, genericSignedBlock.GetCapella())
			result.Cmd = r.Cmd
			newBlock, ok := r.Result.(*ethpb.SignedBeaconBlockCapella)
			if ok {
				genericSignedBlock.Block = &ethpb.GenericSignedBeaconBlock_Capella{
					Capella: newBlock,
				}
				newBlockBase64, _ := common.GenericSignedBlockToBase64(genericSignedBlock)
				result.Result = newBlockBase64
			}
		}
	}

	return result
}

func (s *BlockAPI) AfterPropose(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	genericSignedBlock, err := common.Base64ToGenericSignedBlock(signedBlockDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: signedBlockDataBase64,
		}
	}
	result := types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedBlockDataBase64,
	}

	if t, find := findMaxLevelStrategy(s.b.GetInternalSlotStrategy(), int64(slot)); find {
		action := t.Actions["BlockAfterPropose"]
		if action != nil {
			r := action.RunAction(s.b, int64(slot), pubkey, genericSignedBlock.GetCapella())
			result.Cmd = r.Cmd
			newBlock, ok := r.Result.(*ethpb.SignedBeaconBlockCapella)
			if ok {
				genericSignedBlock.Block = &ethpb.GenericSignedBeaconBlock_Capella{
					Capella: newBlock,
				}
				newBlockBase64, _ := common.GenericSignedBlockToBase64(genericSignedBlock)
				result.Result = newBlockBase64
			}
		}
	}

	return result
}
