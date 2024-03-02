package apis

import (
	"encoding/json"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/plugins"
	"github.com/tsinghua-cel/attacker-service/types"
)

// AttestAPI offers and API for attestation operations.
type AttestAPI struct {
	b      Backend
	plugin plugins.AttackerPlugin
}

// NewAttestAPI creates a new tx pool service that gives information about the transaction pool.
func NewAttestAPI(b Backend, plugin plugins.AttackerPlugin) *AttestAPI {
	return &AttestAPI{b, plugin}
}

func (s *AttestAPI) GetStrategy() []byte {
	d, _ := json.Marshal(s.b.GetStrategy().Attest)
	return d
}

func (s *AttestAPI) UpdateStrategy(data []byte) error {
	var attestStrategy types.AttestStrategy
	if err := json.Unmarshal(data, &attestStrategy); err != nil {
		return err
	}
	s.b.GetStrategy().Attest = attestStrategy
	log.Infof("attest strategy updated to %v\n", attestStrategy)
	return nil
}

func (s *AttestAPI) BeforeBroadCast(slot uint64) types.AttackerResponse {
	if s.plugin != nil {
		result := s.plugin.AttestBeforeBroadCast(pluginContext(s.b), slot)
		return types.AttackerResponse{
			Cmd: result.Cmd,
		}
	}
	return types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}
}

func (s *AttestAPI) AfterBroadCast(slot uint64) types.AttackerResponse {
	if s.plugin != nil {
		result := s.plugin.AttestAfterBroadCast(pluginContext(s.b), slot)
		return types.AttackerResponse{
			Cmd: result.Cmd,
		}
	}

	return types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}
}

func (s *AttestAPI) BeforeSign(slot uint64, pubkey string, attestDataBase64 string) types.AttackerResponse {
	attestation, err := common.Base64ToAttestationData(attestDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: attestDataBase64,
		}
	}

	if s.plugin != nil {
		result := s.plugin.AttestBeforeSign(pluginContext(s.b), slot, pubkey, attestation)
		newAttestation, ok := result.Result.(*ethpb.AttestationData)
		if ok {
			newData, _ := common.AttestationDataToBase64(newAttestation)
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: newData,
			}
		} else {
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: attestDataBase64,
			}
		}
	}

	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: attestDataBase64,
	}
}

func (s *AttestAPI) AfterSign(slot uint64, pubkey string, signedAttestDataBase64 string) types.AttackerResponse {
	signedAttestData, err := common.Base64ToSignedAttestation(signedAttestDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: signedAttestDataBase64,
		}
	}

	if s.plugin != nil {
		result := s.plugin.AttestAfterSign(pluginContext(s.b), slot, pubkey, signedAttestData)
		newAttestation, ok := result.Result.(*ethpb.Attestation)
		if ok {
			newData, _ := common.SignedAttestationToBase64(newAttestation)
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: newData,
			}
		} else {
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: signedAttestDataBase64,
			}
		}
	}

	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedAttestDataBase64,
	}
}

func (s *AttestAPI) BeforePropose(slot uint64, pubkey string, signedAttestDataBase64 string) types.AttackerResponse {
	signedAttest, err := common.Base64ToSignedAttestation(signedAttestDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: signedAttestDataBase64,
		}
	}
	if s.plugin != nil {
		result := s.plugin.AttestBeforePropose(pluginContext(s.b), slot, pubkey, signedAttest)
		newAttestation, ok := result.Result.(*ethpb.Attestation)
		if ok {
			newData, _ := common.SignedAttestationToBase64(newAttestation)
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: newData,
			}
		} else {
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: signedAttestDataBase64,
			}
		}
	}

	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedAttestDataBase64,
	}
}

func (s *AttestAPI) AfterPropose(slot uint64, pubkey string, signedAttestDataBase64 string) types.AttackerResponse {
	signedAttest, err := common.Base64ToSignedAttestation(signedAttestDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: signedAttestDataBase64,
		}
	}
	if s.plugin != nil {
		result := s.plugin.AttestAfterPropose(pluginContext(s.b), slot, pubkey, signedAttest)
		newAttestation, ok := result.Result.(*ethpb.Attestation)
		if ok {
			newData, _ := common.SignedAttestationToBase64(newAttestation)
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: newData,
			}
		} else {
			return types.AttackerResponse{
				Cmd:    result.Cmd,
				Result: signedAttestDataBase64,
			}
		}
	}

	return types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedAttestDataBase64,
	}
}
