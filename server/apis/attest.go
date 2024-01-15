package apis

import (
	"encoding/base64"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

// AttestAPI offers and API for attestation operations.
type AttestAPI struct {
	b Backend
}

// NewBlockAPI creates a new tx pool service that gives information about the transaction pool.
func NewAttestAPI(b Backend) *AttestAPI {
	return &AttestAPI{b}
}

func (s *AttestAPI) ModifySlot(attestDataBase64 string) string {
	attestData, err := base64.StdEncoding.DecodeString(attestDataBase64)
	if err != nil {
		log.WithError(err).Error("base64 decode attest data failed")
		return ""
	}
	var attest = new(ethpb.AttestationData)
	if err := proto.Unmarshal(attestData, attest); err != nil {
		log.WithError(err).Error("unmarshal attest data failed")
		return ""
	}
	modifyAttestData, err := s.internalModifyAttestSlot(attest)
	if err != nil {
		log.WithError(err).Error("modify attest data failed")
		return ""
	}
	return base64.StdEncoding.EncodeToString(modifyAttestData)
}

func (s *AttestAPI) internalModifyAttestSlot(attest *ethpb.AttestationData) ([]byte, error) {
	log.Infof("modify attest slot for attest type %T", attest)
	attest.Slot = attest.Slot + 1
	return proto.Marshal(attest)
}
