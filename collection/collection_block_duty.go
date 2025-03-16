package collection

import (
	"github.com/astaxie/beego/orm"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/dbmodel"
	"strconv"
	"time"
)

var (
	latestBlockDutyEpoch int64 = -1
)

func ScheduleBlockDuty(interval time.Duration, url string) {
	client := beaconapi.NewBeaconGwClient(url)
	orm := orm.NewOrm()
	tc := time.NewTicker(interval)
	defer tc.Stop()
	for {
		select {
		case <-tc.C:
			log.Info("ScheduleBlockDuty")
			if err := GetBlockDutyToMysql(orm, client); err != nil {
				log.WithError(err).Error("ScheduleBlockDuty failed")
			}
		}
	}
}

func GetBlockDutyToMysql(o orm.Ormer, client *beaconapi.BeaconGwClient) error {
	latestHeader, err := client.GetLatestBeaconHeader()
	if err != nil {
		return err
	}

	latestSlot, _ := strconv.ParseInt(latestHeader.Header.Message.Slot, 10, 64)
	curEpoch := common.SlotToEpoch(latestSlot)

	if latestBlockDutyEpoch < 0 {
		latestBlockDutyEpoch = dbmodel.GetMaxBlockDutyEpoch()
	}

	var maxRangeEpoch = 5

	if curEpoch <= latestBlockDutyEpoch {
		return nil
	}

	if (curEpoch - latestBlockDutyEpoch) > int64(maxRangeEpoch) {
		curEpoch = latestBlockDutyEpoch + int64(maxRangeEpoch)
	}

	for epoch := latestBlockDutyEpoch + 1; epoch <= curEpoch; epoch++ {
		duties, err := client.GetProposerDuties(int(epoch))
		if err != nil {
			log.WithField("epoch", epoch).WithError(err).Error("GetBlockDutyToMysql get proposer duties failed")
			return err
		}

		if err := dbmodel.InsertNewBlockDuties(o, epoch, duties); err != nil {
			log.WithField("epoch", epoch).WithError(err).Error("GetBlockDutyToMysql insert proposer duties failed")
			return err
		}
		latestBlockDutyEpoch = epoch
	}

	return nil
}
