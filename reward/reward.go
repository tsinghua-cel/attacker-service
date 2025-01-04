package reward

import (
	"encoding/csv"
	"errors"
	"github.com/astaxie/beego/orm"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/config"
	"github.com/tsinghua-cel/attacker-service/dbmodel"
	"os"
	"strconv"
)

func GetRewardsToMysql(gwEndpoint string) error {
	client := beaconapi.NewBeaconGwClient(gwEndpoint)
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

	curMaxEpoch := dbmodel.GetMaxEpoch()
	epochNumber := curMaxEpoch + 1
	if curMaxEpoch < 0 {
		epochNumber = 0
	}
	o := orm.NewOrm()

	//  开始事务
	if err = o.Begin(); err != nil {
		log.WithError(err).Error("GetRewardsToMysql orm begin failed")
		return err
	}
	repo := dbmodel.NewAttestRewardRepository(o)
	log.WithFields(log.Fields{
		"epochNumber": epochNumber,
		"latestEpoch": latestEpoch,
	}).Debug("GetRewardsToMysql")

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
			if err = repo.Create(record); err != nil {
				o.Rollback()
				return errors.New("insert attest reward failed")
			}
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
				if err = dbmodel.InsertBlockReward(o, record); err != nil {
					o.Rollback()
					return errors.New("insert block reward failed")
				}
			}
		}
		epochNumber++
	}
	if err = o.Commit(); err != nil {
		return errors.New("commit failed")
	}
	return nil
}

func GetRewards(gwEndpoint string, output string) error {
	bakfile := output + ".bak"
	file, err := os.Create(bakfile)
	if err != nil {
		return err
	}
	succeed := false
	defer func() {
		file.Close()
		if succeed {
			os.Rename(bakfile, output)
		}
	}()
	client := beaconapi.NewBeaconGwClient(gwEndpoint)

	slots_per_epoch, err := client.GetIntConfig(beaconapi.SLOTS_PER_EPOCH)
	if err != nil {
		// todo: add log
		return err
	}
	latestHeader, err := client.GetLatestBeaconHeader()
	if err != nil {
		return err
	}

	latestSlot, _ := strconv.ParseInt(latestHeader.Header.Message.Slot, 10, 64)
	latestEpoch := latestSlot / int64(slots_per_epoch)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Epoch", "Validator Index", "Head", "Target", "Source", "Inclusion Delay", "Inactivity"})

	epochNumber := int64(0)

	for epochNumber <= (latestEpoch - 2) {
		info, err := client.GetAllValReward(int(epochNumber))
		if err != nil {
			return err
		}
		for _, totalReward := range info.TotalRewards {
			inclusionDelay := int64(0)
			if totalReward.InclusionDelay != nil {
				inclusionDelay = int64(*totalReward.InclusionDelay)
			}
			writer.Write([]string{
				strconv.FormatInt(epochNumber, 10),
				strconv.FormatInt(int64(totalReward.ValidatorIndex), 10),
				strconv.FormatInt(int64(totalReward.Head), 10),
				strconv.FormatInt(int64(totalReward.Target), 10),
				strconv.FormatInt(int64(totalReward.Source), 10),
				strconv.FormatInt(inclusionDelay, 10),
				strconv.FormatInt(int64(totalReward.Inactivity), 10),
			})
		}

		epochNumber++

	}
	succeed = true
	return err
}
