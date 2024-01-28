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

func (s *BlockAPI) UpdateStrategy(data []byte) error {
	var blockStrategy strategy.BlockStrategy
	if err := json.Unmarshal(data, &blockStrategy); err != nil {
		return err
	}
	s.b.GetStrategy().Block = blockStrategy
	log.Infof("block strategy updated to %v\n", blockStrategy)
	return nil
}

func (s *BlockAPI) BroadCastDelay() types.AttackerResponse {
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

func (s *BlockAPI) modifyBlock(slot uint64, pubkey string, blockDataBase64 string) types.AttackerResponse {
	// 1. 只有每个epoch最后一个出块的恶意节点出块，其他节点不出快
	val := s.b.GetValidatorDataSet().GetValidatorByPubkey(pubkey)
	if val.Role != types.AttackerRole {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: blockDataBase64,
		}
	}

	duties, err := s.b.GetCurrentEpochProposeDuties()
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: blockDataBase64,
		}
	}

	latestAttackerVal := int64(-1)
	for _, duty := range duties {
		dutyValIdx, _ := strconv.Atoi(duty.ValidatorIndex)
		if s.b.GetValidatorRole(dutyValIdx) == types.AttackerRole {
			latestAttackerVal = int64(dutyValIdx)
		}
	}
	if val.Index != latestAttackerVal {
		// 不是最后一个出块的恶意节点，不出块
		return types.AttackerResponse{
			Cmd:    types.CMD_RETURN,
			Result: blockDataBase64,
		}
	}

	genericBlock, err := s.getGenericBlockFromData(blockDataBase64)
	if err != nil {
		log.WithError(err).Error("get block from data failed")
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: blockDataBase64,
		}
	}
	// 2.延迟到下个epoch的中间出块
	block, err := s.getCapellaBlockFromGeneric(genericBlock)
	if err != nil {
		log.WithError(err).Error("get block from data failed")
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: blockDataBase64,
		}
	}
	curSlot := slot
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
	allSlotAttest := s.b.GetAttestSet(uint64(curSlot))
	validatorSet := s.b.GetValidatorDataSet()
	attackerAttestations := make([]*ethpb.Attestation, 0)
	for publicKey, att := range allSlotAttest.Attestations {
		val := validatorSet.GetValidatorByPubkey(publicKey)
		if val != nil && val.Role == types.AttackerRole {
			log.WithField("pubkey", publicKey).Debug("add attacker attestation to block")
			attackerAttestations = append(attackerAttestations, att)
		}
	}
	block.Capella.Body.Attestations = attackerAttestations

	// 4. encode to base64.
	genericBlock.Block = block

	resBlockBase64, err := s.genericBlockToBase64(genericBlock)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: blockDataBase64,
		}
	}
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: resBlockBase64,
	}
}

func (s *BlockAPI) getCapellaBlockFromGenericSigned(block *ethpb.GenericSignedBeaconBlock) (*ethpb.GenericSignedBeaconBlock_Capella, error) {
	switch b := block.Block.(type) {
	case nil:
		return nil, ErrNilObject
	case *ethpb.GenericSignedBeaconBlock_Phase0:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericSignedBeaconBlock_Altair:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericSignedBeaconBlock_Bellatrix:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericSignedBeaconBlock_BlindedBellatrix:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericSignedBeaconBlock_Capella:
		return b, nil
	case *ethpb.GenericSignedBeaconBlock_BlindedCapella:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericSignedBeaconBlock_Deneb:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericSignedBeaconBlock_BlindedDeneb:
		return nil, ErrUnsupportedBeaconBlock
	default:
		log.WithError(ErrUnsupportedBeaconBlock).Errorf("unsupported beacon block from type %T", b)
		return nil, ErrUnsupportedBeaconBlock
	}
}

func (s *BlockAPI) getCapellaBlockFromGeneric(block *ethpb.GenericBeaconBlock) (*ethpb.GenericBeaconBlock_Capella, error) {
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

func (s *BlockAPI) getGenericBlockFromData(blockDataBase64 string) (*ethpb.GenericBeaconBlock, error) {
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
	return block, nil
}

func (s *BlockAPI) getGenericSignedBlockFromData(signedBlockDataBase64 string) (*ethpb.GenericSignedBeaconBlock, error) {
	blockData, err := base64.StdEncoding.DecodeString(signedBlockDataBase64)
	if err != nil {
		log.WithError(err).Error("base64 decode block data failed")
		return nil, err
	}
	var block = new(ethpb.GenericSignedBeaconBlock)
	if err := proto.Unmarshal(blockData, block); err != nil {
		log.WithError(err).Error("unmarshal block data failed")
		return nil, err
	}
	return block, nil
}

func (s *BlockAPI) genericBlockToBase64(block *ethpb.GenericBeaconBlock) (string, error) {
	data, err := proto.Marshal(block)
	if err != nil {
		log.WithError(err).Error("marshal block data failed")
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (s *BlockAPI) genericSignedBlockToBase64(block *ethpb.GenericSignedBeaconBlock) (string, error) {
	data, err := proto.Marshal(block)
	if err != nil {
		log.WithError(err).Error("marshal block data failed")
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (s *BlockAPI) BeforeBroadCast(cliInfo string) types.AttackerResponse {
	return types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}
}

func (s *BlockAPI) AfterBroadCast(cliInfo string) types.AttackerResponse {
	return types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}
}

func (s *BlockAPI) BeforeSign(slot uint64, pubkey string, blockDataBase64 string) types.AttackerResponse {
	modifyBlockRes := s.modifyBlock(slot, pubkey, blockDataBase64)
	return modifyBlockRes
}

func (s *BlockAPI) AfterSign(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedBlockDataBase64,
	}
}

func (s *BlockAPI) BeforePropose(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedBlockDataBase64,
	}
}

func (s *BlockAPI) AfterPropose(slot uint64, pubkey string, signedBlockDataBase64 string) types.AttackerResponse {
	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedBlockDataBase64,
	}

}
