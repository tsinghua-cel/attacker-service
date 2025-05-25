package disguisedRandao

import (
	"github.com/prysmaticlabs/prysm/v5/config/params"
	log "github.com/sirupsen/logrus"
)

func DumpParams() {
	log.WithFields(log.Fields{
		"DomainRandao":                 params.BeaconConfig().DomainRandao,
		"EPOCHS_PER_HISTORICAL_VECTOR": params.BeaconConfig().EpochsPerHistoricalVector,
	}).Info("dump params related to randao")
}
