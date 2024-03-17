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
	strategys := s.b.GetInternalSlotStrategy()
	result := types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}
	for _, t := range strategys {
		if t.Slot.Compare(int64(slot)) == 0 {
			action := t.Actions["AttestBeforeBroadCast"]
			if action != nil {
				r := action.RunAction(s.b, int64(slot), "")
				result.Cmd = r.Cmd
			}
			break
		}
	}

	return result
}

func (s *AttestAPI) AfterBroadCast(slot uint64) types.AttackerResponse {
	strategys := s.b.GetInternalSlotStrategy()
	result := types.AttackerResponse{
		Cmd: types.CMD_NULL,
	}
	for _, t := range strategys {
		if t.Slot.Compare(int64(slot)) == 0 {
			action := t.Actions["AttestAfterBroadCast"]
			if action != nil {
				r := action.RunAction(s.b, int64(slot), "")
				result.Cmd = r.Cmd
			}
			break
		}
	}

	return result
}

func (s *AttestAPI) BeforeSign(slot uint64, pubkey string, attestDataBase64 string) types.AttackerResponse {
	result := types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: attestDataBase64,
	}

	attestation, err := common.Base64ToAttestationData(attestDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: attestDataBase64,
		}
	}

	strategys := s.b.GetInternalSlotStrategy()
	for _, t := range strategys {
		if t.Slot.Compare(int64(slot)) == 0 {
			action := t.Actions["AttestBeforeSign"]
			if action != nil {
				r := action.RunAction(s.b, int64(slot), pubkey, attestation)
				result.Cmd = r.Cmd
				newAttestation, ok := r.Result.(*ethpb.AttestationData)
				if ok {
					newData, _ := common.AttestationDataToBase64(newAttestation)
					result.Result = newData
				}
			}
			break
		}
	}

	return result
}

func (s *AttestAPI) AfterSign(slot uint64, pubkey string, signedAttestDataBase64 string) types.AttackerResponse {
	signedAttestData, err := common.Base64ToSignedAttestation(signedAttestDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: signedAttestDataBase64,
		}
	}
	result := types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedAttestDataBase64,
	}

	strategys := s.b.GetInternalSlotStrategy()
	for _, t := range strategys {
		if t.Slot.Compare(int64(slot)) == 0 {
			action := t.Actions["AttestAfterSign"]
			if action != nil {
				r := action.RunAction(s.b, int64(slot), pubkey, signedAttestData)
				result.Cmd = r.Cmd
				newAttestation, ok := r.Result.(*ethpb.Attestation)
				if ok {
					newData, _ := common.SignedAttestationToBase64(newAttestation)
					result.Result = newData
				}
			}
			break
		}
	}

	return result
}

func (s *AttestAPI) BeforePropose(slot uint64, pubkey string, signedAttestDataBase64 string) types.AttackerResponse {
	signedAttest, err := common.Base64ToSignedAttestation(signedAttestDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: signedAttestDataBase64,
		}
	}
	result := types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedAttestDataBase64,
	}

	strategys := s.b.GetInternalSlotStrategy()
	for _, t := range strategys {
		if t.Slot.Compare(int64(slot)) == 0 {
			action := t.Actions["AttestBeforePropose"]
			if action != nil {
				r := action.RunAction(s.b, int64(slot), pubkey, signedAttest)
				result.Cmd = r.Cmd
				newAttestation, ok := r.Result.(*ethpb.Attestation)
				if ok {
					newData, _ := common.SignedAttestationToBase64(newAttestation)
					result.Result = newData
				}
			}
			break
		}
	}

	return result
}

func (s *AttestAPI) AfterPropose(slot uint64, pubkey string, signedAttestDataBase64 string) types.AttackerResponse {
	signedAttest, err := common.Base64ToSignedAttestation(signedAttestDataBase64)
	if err != nil {
		return types.AttackerResponse{
			Cmd:    types.CMD_NULL,
			Result: signedAttestDataBase64,
		}
	}
	result := types.AttackerResponse{
		Cmd:    types.CMD_NULL,
		Result: signedAttestDataBase64,
	}

	strategys := s.b.GetInternalSlotStrategy()
	for _, t := range strategys {
		if t.Slot.Compare(int64(slot)) == 0 {
			action := t.Actions["AttestAfterPropose"]
			if action != nil {
				r := action.RunAction(s.b, int64(slot), pubkey, signedAttest)
				result.Cmd = r.Cmd
				newAttestation, ok := r.Result.(*ethpb.Attestation)
				if ok {
					newData, _ := common.SignedAttestationToBase64(newAttestation)
					result.Result = newData
				}
			}
			break
		}
	}

	return result
}
