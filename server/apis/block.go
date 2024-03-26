package apis

import (
	"encoding/json"
	"errors"
	"github.com/prysmaticlabs/prysm/v5/cache/lru"
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
	return s.todoActionsWithSlot(slot, "BlockDelayForBroadCast")
}

func (s *BlockAPI) DelayForReceiveBlock(slot uint64) types.AttackerResponse {
	return s.todoActionsWithSlot(slot, "BlockDelayForReceiveBlock")
}

func (s *BlockAPI) BeforeBroadCast(slot uint64) types.AttackerResponse {
	return s.todoActionsWithSlot(slot, "BlockBeforeBroadCast")
}

func (s *BlockAPI) AfterBroadCast(slot uint64) types.AttackerResponse {
	return s.todoActionsWithSlot(slot, "BlockAfterBroadCast")
}

func (s *BlockAPI) BeforeSign(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	return s.todoActionsWithSignedBlock(slot, pubkey, signedBlockDataBase64, "BlockBeforeSign")
}

func (s *BlockAPI) AfterSign(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	return s.todoActionsWithSignedBlock(slot, pubkey, signedBlockDataBase64, "BlockAfterSign")
}

func (s *BlockAPI) BeforePropose(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	return s.todoActionsWithSignedBlock(slot, pubkey, signedBlockDataBase64, "BlockBeforePropose")
}

func (s *BlockAPI) AfterPropose(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	return s.todoActionsWithSignedBlock(slot, pubkey, signedBlockDataBase64, "BlockAfterPropose")
}

func (s *BlockAPI) todoActionsWithSlot(slot uint64, name string) types.AttackerResponse {
	result := types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}

	if t, find := findMaxLevelStrategy(s.b.GetInternalSlotStrategy(), int64(slot)); find {
		action := t.Actions[name]
		if action != nil {
			r := action.RunAction(s.b, int64(slot), "")
			result.Cmd = r.Cmd
		}
	}
	log.WithFields(log.Fields{
		"cmd":    result.Cmd,
		"slot":   slot,
		"action": name,
	}).Info("exit todoActionsWithSlot")

	return result
}

func (s *BlockAPI) todoActionsWithSignedBlock(slot uint64, pubkey string, signedBlockDataBase64 string, name string) types.AttackerResponse {
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
		action := t.Actions[name]
		if action != nil {
			block, err := common.GetDenebBlockFromGenericSignedBlock(genericSignedBlock)
			if err != nil {
				log.WithError(err).WithField("slot", slot).Error("get block instance failed")
				return result
			}
			r := action.RunAction(s.b, int64(slot), pubkey, block)
			result.Cmd = r.Cmd
			if newBlockBase64, err := common.GenericSignedBlockToBase64(genericSignedBlock); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"slot":   slot,
					"action": name,
				}).Error("marshal to block failed")
			} else {
				result.Result = newBlockBase64
			}
		}
	}
	log.WithFields(log.Fields{
		"cmd":    result.Cmd,
		"slot":   slot,
		"action": name,
	}).Info("exit todoActionsWithBlock")

	return result
}
