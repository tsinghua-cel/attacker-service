package apis

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/strategy"
	"github.com/tsinghua-cel/attacker-service/types"
	"google.golang.org/protobuf/proto"
	"strconv"
	"time"
)

var (
	ErrNilObject              = errors.New("nil object")
	ErrUnsupportedBeaconBlock = errors.New("unsupported beacon block")
)

// BlockAPI offers and API for block operations.
type BlockAPI struct {
	b Backend
}

// NewBlockAPI creates a new tx pool service that gives information about the transaction pool.
func NewBlockAPI(b Backend) *BlockAPI {
	return &BlockAPI{b}
}

func (s *BlockAPI) GetStrategy(cliInfo string) []byte {
	d, _ := json.Marshal(s.b.GetStrategy().Block)
	return d
}

func (s *BlockAPI) UpdateStrategy(cliInfo string, data []byte) error {
	var blockStrategy strategy.BlockStrategy
	if err := json.Unmarshal(data, &blockStrategy); err != nil {
		return err
	}
	s.b.GetStrategy().Block = blockStrategy
	log.Infof("block strategy updated to %v\n", blockStrategy)
	return nil
}

func (s *BlockAPI) BroadCastDelay(cliInfo string) types.AttackerResponse {
	bs := s.b.GetStrategy().Block
	if !bs.DelayEnable {
		return types.AttackerResponse{
			Cmd: types.CMD_NULL,
		}
	}
	time.Sleep(time.Millisecond * time.Duration(s.b.GetStrategy().Block.BroadCastDelay))
	return types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}
}

func (s *BlockAPI) ModifyBlock(cliInfo string, blockDataBase64 string) types.AttackerResponse {
	// 1. 只有每个epoch最后一个出块的恶意节点出块，其他节点不出快
	cInfo := types.ToClientInfo(cliInfo)
	duties, err := s.b.GetCurrentEpochProposeDuties()
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: blockDataBase64,
		}
	}
	latestAttackerVal := -1
	for _, duty := range duties {
		dutyValIdx, _ := strconv.Atoi(duty.ValidatorIndex)
		if s.b.GetValidatorRole(dutyValIdx) == types.AttackerRole {
			latestAttackerVal = dutyValIdx
		}
	}
	if cInfo.ValidatorIndex != latestAttackerVal {
		// 不是最后一个出块的恶意节点，不出块
		return types.AttackerResponse{
			Cmd:    types.CMD_RETURN,
			Result: blockDataBase64,
		}
	}
	// 2.延迟到下个epoch的中间出块
	block, err := s.getCapellaBlockFromData(blockDataBase64)
	if err != nil {
		log.WithError(err).Error("get block from data failed")
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: blockDataBase64,
		}
	}
	curSlot := block.Capella.Slot
	slotsPerEpoch := s.b.GetSlotsPerEpoch()
	secondsPerSlot := s.b.GetIntervalPerSlot()
	curEpoch := uint64(curSlot) / uint64(slotsPerEpoch)

	intervalSlots := uint64(slotsPerEpoch)*(curEpoch+1) - 1 + uint64(slotsPerEpoch)/2 - uint64(curSlot)
	intervalSeconds := int(intervalSlots) * secondsPerSlot
	log.WithFields(log.Fields{
		"intervalSlots":   intervalSlots,
		"intervalSeconds": intervalSeconds,
	})
	time.Sleep(time.Second * time.Duration(intervalSeconds))

	// 3.出的块的一个字段attestation要包含其他恶意节点的attestation。
	attestDuties, err := s.b.Get
	// todo: get other attacker node attestation.
	// 这里指的是 加上其他恶意节点在当前epoch内的attestation 吗？
	block.Capella.Body.Attestations

}

func (s *BlockAPI) ModifySlot(cliInfo string, blockDataBase64 string) types.AttackerResponse {
	blockData, err := base64.StdEncoding.DecodeString(blockDataBase64)
	if err != nil {
		log.WithError(err).Error("base64 decode block data failed")
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: blockDataBase64,
		}
	}
	var block = new(ethpb.GenericBeaconBlock)
	if err := proto.Unmarshal(blockData, block); err != nil {
		log.WithError(err).Error("unmarshal block data failed")
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: blockDataBase64,
		}
	}
	modifyBlockData, err := s.internalModifyBlockSlot(block)
	if err != nil {
		log.WithError(err).Error("modify block data failed")
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: blockDataBase64,
		}
	}
	ndata := base64.StdEncoding.EncodeToString(modifyBlockData)
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: ndata,
	}
}

// implement every modify function
func (s *BlockAPI) modifyBlockFromProtoPhase0(block *ethpb.BeaconBlock) {
	block.Slot = block.Slot + 1

}

func (s *BlockAPI) modifyBlockFromProtoAltair(block *ethpb.BeaconBlockAltair) {
	block.Slot = block.Slot + 1
}

func (s *BlockAPI) modifyBlockFromProtoBellatrix(block *ethpb.BeaconBlockBellatrix) {
	block.Slot = block.Slot + 1

}

func (s *BlockAPI) modifyBlindedBlockFromProtoBellatrix(block *ethpb.BlindedBeaconBlockBellatrix) {
	block.Slot = block.Slot + 1
}

func (s *BlockAPI) modifyBlockFromProtoCapella(block *ethpb.BeaconBlockCapella) {
	block.Slot = block.Slot + 1
}

func (s *BlockAPI) modifyBlindedBlockFromProtoCapella(block *ethpb.BlindedBeaconBlockCapella) {
	block.Slot = block.Slot + 1
}

func (s *BlockAPI) modifyBlockFromProtoDeneb(block *ethpb.BeaconBlockDeneb) {
	block.Slot = block.Slot + 1
}

func (s *BlockAPI) modifyBlindedBlockFromProtoDeneb(block *ethpb.BlindedBeaconBlockDeneb) {
	block.Slot = block.Slot + 1
}

func (s *BlockAPI) getCapellaBlockFromData(blockDataBase64 string) (*ethpb.GenericBeaconBlock_Capella, error) {
	blockData, err := base64.StdEncoding.DecodeString(blockDataBase64)
	if err != nil {
		log.WithError(err).Error("base64 decode block data failed")
		return nil, err
	}
	var block = new(ethpb.GenericBeaconBlock)
	if err := proto.Unmarshal(blockData, block); err != nil {
		log.WithError(err).Error("unmarshal block data failed")
		return nil, err
	}

	switch b := block.Block.(type) {
	case nil:
		return nil, ErrNilObject
	case *ethpb.GenericBeaconBlock_Phase0:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericBeaconBlock_Altair:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericBeaconBlock_Bellatrix:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericBeaconBlock_BlindedBellatrix:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericBeaconBlock_Capella:
		return b, nil
	case *ethpb.GenericBeaconBlock_BlindedCapella:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericBeaconBlock_Deneb:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericBeaconBlock_BlindedDeneb:
		return nil, ErrUnsupportedBeaconBlock
	default:
		log.WithError(ErrUnsupportedBeaconBlock).Errorf("unsupported beacon block from type %T", b)
		return nil, ErrUnsupportedBeaconBlock
	}
}

func (s *BlockAPI) internalModifyBlockSlot(blk *ethpb.GenericBeaconBlock) ([]byte, error) {
	log.Infof("modify block slot for blk type %T", blk.Block)
	switch b := blk.Block.(type) {
	case nil:
		return nil, ErrNilObject
	case *ethpb.GenericBeaconBlock_Phase0:
		s.modifyBlockFromProtoPhase0(b.Phase0)
	case *ethpb.GenericBeaconBlock_Altair:
		s.modifyBlockFromProtoAltair(b.Altair)
	case *ethpb.GenericBeaconBlock_Bellatrix:
		s.modifyBlockFromProtoBellatrix(b.Bellatrix)
	case *ethpb.GenericBeaconBlock_BlindedBellatrix:
		s.modifyBlindedBlockFromProtoBellatrix(b.BlindedBellatrix)
	case *ethpb.GenericBeaconBlock_Capella:
		s.modifyBlockFromProtoCapella(b.Capella)
	case *ethpb.GenericBeaconBlock_BlindedCapella:
		s.modifyBlindedBlockFromProtoCapella(b.BlindedCapella)
	case *ethpb.GenericBeaconBlock_Deneb:
		s.modifyBlockFromProtoDeneb(b.Deneb.Block)
	//case *ethpb.BlindedBeaconBlockDeneb:
	//	s.modifyBlindedBlockFromProtoDeneb(b)
	//case *ethpb.GenericBeaconBlock_BlindedDeneb:
	//	s.modifyBlindedBlockFromProtoDeneb(b.BlindedDeneb.Block)
	default:
		log.WithError(ErrUnsupportedBeaconBlock).Errorf("unsupported beacon block from type %T", b)
		return nil, ErrUnsupportedBeaconBlock
	}
	return proto.Marshal(blk)
}

func (s *BlockAPI) BeforeBroadCast() types.AttackerResponse {
	return types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}

}

func (s *BlockAPI) AfterBroadCast() types.AttackerResponse {
	return types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}
}

func (s *BlockAPI) BeforeSign(cliInfo string, slot uint64, pubkey string, blockDataBase64 string) types.AttackerResponse {
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: blockDataBase64,
	}
}

func (s *BlockAPI) AfterSign(cliInfo string, slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedBlockDataBase64,
	}
}

func (s *BlockAPI) BeforePropose(cliInfo string, slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedBlockDataBase64,
	}
}

func (s *BlockAPI) AfterPropose(cliInfo string, slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedBlockDataBase64,
	}

}
