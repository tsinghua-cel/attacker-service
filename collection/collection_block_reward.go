package collection

import (
	"github.com/astaxie/beego/orm"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"github.com/tsinghua-cel/attacker-service/dbmodel"
	"strconv"
	"time"
)

var (
	latestBlockRewardSlot int64 = -1
)

func ScheduleBlockReward(interval time.Duration, url string) {
	client := beaconapi.NewBeaconGwClient(url)
	orm := orm.NewOrm()
	tc := time.NewTicker(interval)
	defer tc.Stop()
	for {
		select {
		case <-tc.C:
			log.Info("ScheduleBlockReward")
			if err := GetBlockRewardsToMysql(orm, client); err != nil {
				log.WithError(err).Error("ScheduleBlockReward failed")
			}
		}
	}
}

func GetBlockRewardsToMysql(o orm.Ormer, client *beaconapi.BeaconGwClient) error {
	latestHeader, err := client.GetLatestBeaconHeader()
	if err != nil {
		return err
	}

	latestSlot, _ := strconv.ParseInt(latestHeader.Header.Message.Slot, 10, 64)

	if latestBlockRewardSlot < 0 {
		latestBlockRewardSlot = dbmodel.GetMaxBlockRewardSlot(o)
	}

	var maxRangeSlot = 32
	var safeSlotGenerate = 64 // two epoch

	if latestSlot <= (latestBlockRewardSlot + int64(safeSlotGenerate)) {
		return nil
	}
	latestSlot = latestSlot - int64(safeSlotGenerate)

	if (latestSlot - latestBlockRewardSlot) > int64(maxRangeSlot) {
		latestSlot = latestBlockRewardSlot + int64(maxRangeSlot)
	}

	var blkRewardInfo = make([]*dbmodel.BlockReward, 0)

	for slot := latestBlockRewardSlot + 1; slot <= latestSlot; slot++ {
		blockReward, err := client.GetBlockReward(int(slot))
		if err != nil {
			log.WithField("slot", slot).WithError(err).Error("GetBlockRewardsToMysql get block rewards failed, ignore")
			continue
		}
		proposerIdx := blockReward.ProposerIndex
		totalAmount := blockReward.Total
		attestationAmount := blockReward.Attestations
		syncAggregateAmount := blockReward.SyncAggregate
		proposerSlashingsAmount := blockReward.ProposerSlashings
		attesterSlashingsAmount := blockReward.AttesterSlashings
		record := &dbmodel.BlockReward{
			Slot:                   slot,
			ProposerIndex:          int(proposerIdx),
			TotalAmount:            int64(totalAmount),
			AttestationAmount:      int64(attestationAmount),
			SyncAggregateAmount:    int64(syncAggregateAmount),
			ProposerSlashingAmount: int64(proposerSlashingsAmount),
			AttesterSlashingAmount: int64(attesterSlashingsAmount),
		}
		blkRewardInfo = append(blkRewardInfo, record)
	}
	if err := dbmodel.InsertBlockRewardList(o, blkRewardInfo); err != nil {
		log.WithField("slot", latestSlot).WithError(err).Error("GetBlockRewardsToMysql insert block rewards failed")
		return err
	}
	latestBlockRewardSlot = latestSlot
	return nil
}
