package apis

import (
	"encoding/base64"
	"errors"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
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

func (s *BlockAPI) ModifySlot(blockDataBase64 string) string {
	blockData, err := base64.StdEncoding.DecodeString(blockDataBase64)
	if err != nil {
		log.WithError(err).Error("base64 decode block data failed")
		return ""
	}
	var block = new(ethpb.GenericBeaconBlock)
	if err := proto.Unmarshal(blockData, block); err != nil {
		log.WithError(err).Error("unmarshal block data failed")
		return ""
	}
	modifyBlockData, err := s.internalModifyBlockSlot(block)
	if err != nil {
		log.WithError(err).Error("modify block data failed")
		return ""
	}
	return base64.StdEncoding.EncodeToString(modifyBlockData)
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
