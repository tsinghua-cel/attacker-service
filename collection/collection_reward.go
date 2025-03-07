package collection

import (
	"github.com/astaxie/beego/orm"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/config"
	"github.com/tsinghua-cel/attacker-service/dbmodel"
	"strconv"
	"sync"
)

var (
	once     sync.Once
	localorm orm.Ormer
)

func getOrm() orm.Ormer {
	once.Do(func() {
		localorm = orm.NewOrm()
	})
	return localorm
}

func GetRewardsToMysql(client *beaconapi.BeaconGwClient) error {
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

	curMaxEpoch := dbmodel.GetMaxRewardedEpoch()
	epochNumber := curMaxEpoch + 1
	if curMaxEpoch < 0 {
		epochNumber = 0
	}
	o := getOrm()

	log.WithFields(log.Fields{
		"epochNumber": epochNumber,
		"latestEpoch": latestEpoch,
	}).Debug("GetRewardsToMysql")

	var attRewardInfo = make([]*dbmodel.AttestReward, 0)
	var blkRewardInfo = make([]*dbmodel.BlockReward, 0)

	safeInterval := config.GetSafeEpochEndInterval()
	for (latestEpoch - epochNumber) >= safeInterval {
		info, err := client.GetAllValReward(int(epochNumber))
		if err != nil {
			return err
		}
		for _, totalReward := range info.TotalRewards {
			valIdx := totalReward.ValidatorIndex
			headAmount := int64(totalReward.Head)
			targetAmount := int64(totalReward.Target)
			sourceAmount := int64(totalReward.Source)
			record := &dbmodel.AttestReward{
				Epoch:          epochNumber,
				ValidatorIndex: int(valIdx),
				HeadAmount:     headAmount,
				TargetAmount:   targetAmount,
				SourceAmount:   sourceAmount,
			}
			attRewardInfo = append(attRewardInfo, record)
		}

		// get block reward for each slot
		epochStart := common.EpochStart(epochNumber)
		epochEnd := common.EpochEnd(epochNumber)
		for slot := epochStart; slot <= epochEnd; slot++ {
			blockReward, err := client.GetBlockReward(int(slot))
			if err != nil {
				continue
			}
			{

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
		}
		epochNumber++
	}
	if err := dbmodel.InsertBlockRewardList(o, blkRewardInfo); err != nil {
		log.WithError(err).Error("GetRewardsToMysql insert block rewards failed")
		//return err
	}
	if err := dbmodel.InsertAttestRewardList(o, attRewardInfo); err != nil {
		log.WithError(err).Error("GetRewardsToMysql insert attester rewards failed")
		//return err
	}
	return nil
}
