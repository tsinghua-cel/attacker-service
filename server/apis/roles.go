package apis

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/common"
)

// RoleAPI offers and API for role operations.
type AdminAPI struct {
	b Backend
}

// NewRoleAPI creates a new tx pool service that gives information about the transaction pool.
func NewAdminAPI(b Backend) *AdminAPI {
	return &AdminAPI{b}
}

func (s *AdminAPI) SetRoleAttacker(valIndex int) {
	//valSet := s.b.GetValidatorDataSet()
	//valSet.SetValidatorRole(valIndex, types.AttackerRole)
}

func (s *AdminAPI) SetRoleNormal(valIndex int) {
	//valSet := s.b.GetValidatorDataSet()
	//valSet.SetValidatorRole(valIndex, types.NormalRole)
}

func (s *AdminAPI) CommitValidatorsKeys(pubkeys []string, privates []string) error {
	// store all keys for validators.
	return s.b.CommitValidatorsKeys(pubkeys, privates)
}

func (s *AdminAPI) CommitReceivedAttestation(signedAttestDataBase64 string) error {
	// commit the received attestation.
	signedAttestData, err := common.Base64ToSignedAttestation(signedAttestDataBase64)
	if err != nil {
		return fmt.Errorf("failed to decode signed attestation: %w", err)
	}
	s.b.AddAttestToPool(0, "", signedAttestData)
	return nil
}
