package apis

import (
	"encoding/base64"
	"encoding/json"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/strategy"
	"google.golang.org/protobuf/proto"
	"time"
)

// AttestAPI offers and API for attestation operations.
type AttestAPI struct {
	b Backend
}

// NewBlockAPI creates a new tx pool service that gives information about the transaction pool.
func NewAttestAPI(b Backend) *AttestAPI {
	return &AttestAPI{b}
}

func (s *AttestAPI) GetStrategy() []byte {
	d, _ := json.Marshal(s.b.GetStrategy().Attest)
	return d
}

func (s *AttestAPI) UpdateStrategy(data []byte) error {
	var attestStrategy strategy.AttestStrategy
	if err := json.Unmarshal(data, &attestStrategy); err != nil {
		return err
	}
	s.b.GetStrategy().Attest = attestStrategy
	log.Infof("attest strategy updated to %v\n", attestStrategy)
	return nil
}

func (s *AttestAPI) BroadCastDelay() {
	as := s.b.GetStrategy().Attest
	if !as.DelayEnable {
		return
	}
	time.Sleep(time.Millisecond * time.Duration(s.b.GetStrategy().Attest.BroadCastDelay))
}

func (s *AttestAPI) ModifyAttest(slot int64, pubkey string, attestDataBase64 string) string {
	as := s.b.GetStrategy().Attest
	if !as.ModifyEnable {
		return attestDataBase64
	}

	attestData, err := base64.StdEncoding.DecodeString(attestDataBase64)
	if err != nil {
		log.WithError(err).Error("base64 decode attest data failed")
		return attestDataBase64
	}
	var attest = new(ethpb.AttestationData)
	if err := proto.Unmarshal(attestData, attest); err != nil {
		log.WithError(err).Error("unmarshal attest data failed")
		return attestDataBase64
	}

	var (
		modifyAttestData []byte
	)

	// this is a simple case to modify attest.Slot value.
	// you can implement case what you want to do.

	for {

		// and you can do some condition check from execute-node.
		if height, err := s.b.GetBlockHeight(); err == nil {
			if height%2 == 0 {
				break
			}
		}

		modifyAttestData, err = s.internalModifyAttestSlot(attest)
		if err != nil {
			log.WithError(err).Error("modify attest data failed")
		}

		break
	}

	if err != nil || len(modifyAttestData) == 0 {
		// return origin attest data when error occur.
		modifyAttestData = attestData
	}

	return base64.StdEncoding.EncodeToString(modifyAttestData)
}

func (s *AttestAPI) internalModifyAttestSlot(attest *ethpb.AttestationData) ([]byte, error) {
	log.Infof("modify attest slot for attest type %T", attest)
	attest.Slot = attest.Slot + 1
	return proto.Marshal(attest)
}
