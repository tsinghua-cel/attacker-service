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
	latestAttestRewardEpoch int64 = -1
)

func LatestAttestRewardEpoch() int64 {
	return latestAttestRewardEpoch
}

func ScheduleAttestReward(interval time.Duration, url string) {
	client := beaconapi.NewBeaconGwClient(url)
	orm := orm.NewOrm()
	tc := time.NewTicker(interval)
	defer tc.Stop()
	for {
		select {
		case <-tc.C:
			log.Info("ScheduleAttestReward")
			if err := GetAttestRewardsToMysql(orm, client); err != nil {
				log.WithError(err).Error("ScheduleAttestReward failed")
			}
		}
	}
}

func GetAttestRewardsToMysql(o orm.Ormer, client *beaconapi.BeaconGwClient) error {
	latestHeader, err := client.GetLatestBeaconHeader()
	if err != nil {
		return err
	}

	latestSlot, _ := strconv.ParseInt(latestHeader.Header.Message.Slot, 10, 64)
	curEpoch := common.SlotToEpoch(latestSlot)

	if latestAttestRewardEpoch < 0 {
		latestAttestRewardEpoch = dbmodel.GetMaxAttestRewardEpoch(o)
	}

	var maxRangeEpoch = 5
	var safeEpochGenerate = 2

	if curEpoch <= (latestAttestRewardEpoch + int64(safeEpochGenerate)) {
		return nil
	}
	curEpoch = curEpoch - int64(safeEpochGenerate)

	if (curEpoch - latestAttestRewardEpoch) > int64(maxRangeEpoch) {
		curEpoch = latestAttestRewardEpoch + int64(maxRangeEpoch)
	}

	for epoch := latestAttestRewardEpoch + 1; epoch <= curEpoch; epoch++ {
		info, err := client.GetAllValReward(int(epoch))
		if err != nil {
			log.WithField("epoch", epoch).WithError(err).Error("GetAttestRewardsToMysql get attester rewards failed")
			return err
		}
		var attRewardInfo = make([]*dbmodel.AttestReward, 0)
		for _, totalReward := range info.TotalRewards {
			valIdx := totalReward.ValidatorIndex
			headAmount := int64(totalReward.Head)
			targetAmount := int64(totalReward.Target)
			sourceAmount := int64(totalReward.Source)
			record := &dbmodel.AttestReward{
				Epoch:          epoch,
				ValidatorIndex: int(valIdx),
				HeadAmount:     headAmount,
				TargetAmount:   targetAmount,
				SourceAmount:   sourceAmount,
			}
			attRewardInfo = append(attRewardInfo, record)
		}
		if err := dbmodel.InsertAttestRewardList(o, attRewardInfo); err != nil {
			log.WithField("epoch", epoch).WithError(err).Error("GetAttestRewardsToMysql insert attester rewards failed")
			return err
		}
		latestAttestRewardEpoch = epoch
	}
	return nil
}
