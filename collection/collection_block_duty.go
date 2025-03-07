package collection

import (
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"github.com/tsinghua-cel/attacker-service/dbmodel"
	"strconv"
)

func GetBlockDutyToMysql(client *beaconapi.BeaconGwClient) error {
	slots_per_epoch, err := client.GetIntConfig(beaconapi.SLOTS_PER_EPOCH)
	if err != nil {
		log.WithError(err).Error("GetRewardsToMysql get chain config failed")
		return err
	}
	latestHeader, err := client.GetLatestBeaconHeader()
	if err != nil {
		return err
	}

	latestSlot, _ := strconv.ParseInt(latestHeader.Header.Message.Slot, 10, 64)
	latestEpoch := latestSlot / int64(slots_per_epoch)

	curMaxEpoch := dbmodel.GetMaxBlockDutyEpoch()
	if latestEpoch <= curMaxEpoch {
		return nil
	}
	epochNumber := curMaxEpoch + 1

	duties, err := client.GetProposerDuties(int(epochNumber))
	if err != nil {
		log.WithError(err).Error("GetBlockDutyToMysql get proposer duties failed")
		return err
	}
	if len(duties) == 0 {
		log.WithField("epoch", epochNumber).Info("no proposer duties")
		return nil
	}

	o := getOrm()

	if err := dbmodel.InsertNewBlockDuties(o, epochNumber, duties); err != nil {
		log.WithError(err).Error("GetBlockDutyToMysql insert proposer duties failed")
		return err
	}
	return nil
}
