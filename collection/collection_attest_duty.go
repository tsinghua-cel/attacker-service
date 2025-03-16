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
	latestAttestDutyEpoch int64 = -1
)

func ScheduleAttestDuty(interval time.Duration, url string) {
	client := beaconapi.NewBeaconGwClient(url)
	orm := orm.NewOrm()
	tc := time.NewTicker(interval)
	defer tc.Stop()
	for {
		select {
		case <-tc.C:
			log.Info("ScheduleAttestDuty")
			if err := GetAttestDutyToMysql(orm, client); err != nil {
				log.WithError(err).Error("ScheduleAttestDuty failed")
			}
		}
	}
}

func GetAttestDutyToMysql(o orm.Ormer, client *beaconapi.BeaconGwClient) error {
	latestHeader, err := client.GetLatestBeaconHeader()
	if err != nil {
		return err
	}

	latestSlot, _ := strconv.ParseInt(latestHeader.Header.Message.Slot, 10, 64)
	curEpoch := common.SlotToEpoch(latestSlot)

	if latestAttestDutyEpoch < 0 {
		latestAttestDutyEpoch = dbmodel.GetMaxAttestDutyEpoch()
	}

	var maxRangeEpoch = 5

	if curEpoch <= latestAttestDutyEpoch {
		return nil
	}

	if (curEpoch - latestAttestDutyEpoch) > int64(maxRangeEpoch) {
		curEpoch = latestAttestDutyEpoch + int64(maxRangeEpoch)
	}

	for epoch := latestAttestDutyEpoch + 1; epoch <= curEpoch; epoch++ {
		duties, err := client.GetAttesterDuties(int(epoch), []int{})
		if err != nil {
			log.WithField("epoch", epoch).WithError(err).Error("GetAttestDutyToMysql get attester duties failed")
			return err
		}

		if err := dbmodel.InsertNewAttestDuties(o, epoch, duties); err != nil {
			log.WithField("epoch", epoch).WithError(err).Error("GetAttestDutyToMysql insert attester duties failed")
			return err
		}
		latestAttestDutyEpoch = epoch
	}

	return nil
}
