package common

import (
	"bytes"
	"encoding/base64"
	"errors"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var (
	ErrNilObject              = errors.New("nil object")
	ErrUnsupportedBeaconBlock = errors.New("unsupported beacon block")
)

const SSZPrefix = "SSZ:"

func Base64ToAttestationData(attestDataBase64 string) (*ethpb.AttestationData, error) {
	attestData, err := base64.StdEncoding.DecodeString(attestDataBase64)
	if err != nil {
		log.WithError(err).Error("base64 decode attest data failed")
		return nil, err
	}
	var attestation = new(ethpb.AttestationData)
	// detect SSZ prefix (must be uppercase)
	if len(attestData) >= len(SSZPrefix) && bytes.Equal(attestData[:len(SSZPrefix)], []byte(SSZPrefix)) {
		// SSZ encoded payload after prefix
		payload := attestData[len(SSZPrefix):]
		if err := attestation.UnmarshalSSZ(payload); err != nil {
			log.WithError(err).Error("ssz unmarshal attest data failed")
			return nil, err
		}
		return attestation, nil
	} else {
		// use protobuf unmarshal
		if err := proto.Unmarshal(attestData, attestation); err != nil {
			log.WithError(err).Error("unmarshal attest data failed")
			return nil, err
		}
	}

	return attestation, nil
}

func AttestationDataToBase64(attestation *ethpb.AttestationData, useSSZ bool) (string, error) {
	if attestation == nil {
		return "", ErrNilObject
	}
	var data []byte
	var err error
	if useSSZ {
		data, err = attestation.MarshalSSZ()
		if err != nil {
			log.WithError(err).Error("ssz marshal attest data failed")
			return "", err
		}
		// prefix with SSZ marker
		data = append([]byte(SSZPrefix), data...)
	} else {
		data, err = proto.Marshal(attestation)
		if err != nil {
			log.WithError(err).Error("marshal attest data failed")
			return "", err
		}
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func Base64ToSignedAttestation(signedAttestDataBase64 string) (*ethpb.Attestation, error) {
	signedAttestData, err := base64.StdEncoding.DecodeString(signedAttestDataBase64)
	if err != nil {
		log.WithError(err).Error("base64 decode signed attest data failed")
		return nil, err
	}
	var signedAttestation = new(ethpb.Attestation)
	if len(signedAttestData) >= len(SSZPrefix) && bytes.Equal(signedAttestData[:len(SSZPrefix)], []byte(SSZPrefix)) {
		payload := signedAttestData[len(SSZPrefix):]

		if err := signedAttestation.UnmarshalSSZ(payload); err != nil {
			log.WithError(err).Error("ssz unmarshal signed attest data failed")
			return nil, err
		}
		return signedAttestation, nil
	}
	if err := proto.Unmarshal(signedAttestData, signedAttestation); err != nil {
		log.WithError(err).Error("unmarshal signed attest data failed")
		return nil, err
	}
	return signedAttestation, nil
}

func SignedAttestationToBase64(signedAttestation *ethpb.Attestation, useSSZ bool) (string, error) {
	if signedAttestation == nil {
		return "", ErrNilObject
	}
	var data []byte
	var err error
	if useSSZ {
		data, err = signedAttestation.MarshalSSZ()
		if err != nil {
			log.WithError(err).Error("ssz marshal signed attest data failed")
			return "", err
		}
		data = append([]byte(SSZPrefix), data...)
	} else {
		data, err = proto.Marshal(signedAttestation)
		if err != nil {
			log.WithError(err).Error("marshal signed attest data failed")
			return "", err
		}
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func Base64ToSignedDenebBlock(signedBlockBase64 string) (*ethpb.SignedBeaconBlockDeneb, error) {
	signedBlockData, err := base64.StdEncoding.DecodeString(signedBlockBase64)
	if err != nil {
		log.WithError(err).Error("base64 decode signed block data failed")
		return nil, err
	}
	var signedBlock = new(ethpb.SignedBeaconBlockDeneb)
	if len(signedBlockData) >= len(SSZPrefix) && bytes.Equal(signedBlockData[:len(SSZPrefix)], []byte(SSZPrefix)) {
		payload := signedBlockData[len(SSZPrefix):]
		if err := signedBlock.UnmarshalSSZ(payload); err != nil {
			log.WithError(err).Error("ssz unmarshal signed block data failed")
			return nil, err
		}
		return signedBlock, nil
	}
	if err := proto.Unmarshal(signedBlockData, signedBlock); err != nil {
		log.WithError(err).Error("unmarshal signed block data failed")
		return nil, err
	}
	return signedBlock, nil
}

func SignedDenebBlockToBase64(signedBlock *ethpb.SignedBeaconBlockDeneb, useSSZ bool) (string, error) {
	if signedBlock == nil {
		return "", ErrNilObject
	}
	var data []byte
	var err error
	if useSSZ {
		data, err = signedBlock.MarshalSSZ()
		if err != nil {
			log.WithError(err).Error("ssz marshal signed block data failed")
			return "", err
		}
		data = append([]byte(SSZPrefix), data...)
	} else {
		data, err = proto.Marshal(signedBlock)
		if err != nil {
			log.WithError(err).Error("marshal signed block data failed")
			return "", err
		}
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func GetDenebBlockFromGenericSignedBlock(signedBlock *ethpb.GenericSignedBeaconBlock) (*ethpb.SignedBeaconBlockDeneb, error) {
	if signedBlock == nil {
		return nil, ErrNilObject
	}
	switch b := signedBlock.Block.(type) {
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
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericSignedBeaconBlock_BlindedCapella:
		return nil, ErrUnsupportedBeaconBlock
	case *ethpb.GenericSignedBeaconBlock_Deneb:
		return b.Deneb.Block, ErrUnsupportedBeaconBlock
	case *ethpb.GenericSignedBeaconBlock_BlindedDeneb:
		return nil, ErrUnsupportedBeaconBlock
	default:
		log.WithError(ErrUnsupportedBeaconBlock).Errorf("unsupported beacon block from type %T", b)
		return nil, ErrUnsupportedBeaconBlock
	}
}

func GetCapellaBlockFromGenericSignedBlock(signedBlock *ethpb.GenericSignedBeaconBlock) (*ethpb.SignedBeaconBlockCapella, error) {
	if signedBlock == nil {
		return nil, ErrNilObject
	}
	switch b := signedBlock.Block.(type) {
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
		return b.Capella, nil
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
