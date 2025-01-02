package staircase

import (
	"context"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
	"time"
)

type Instance struct {
}

func (o *Instance) Run(ctx context.Context, params types.LibraryParams, feedbacker types.FeedBacker) {
	log.WithField("name", o.Name()).Info("start to run strategy")
	var latestEpoch int64 = -1
	ticker := time.NewTicker(time.Second * 3)
	attacker := params.Attacker
	for {
		select {
		case <-ctx.Done():
			log.WithField("name", o.Name()).Info("stop to run strategy")
			return
		case <-ticker.C:
			slot := attacker.GetCurSlot()
			log.WithFields(log.Fields{
				"slot":      slot,
				"lastEpoch": latestEpoch,
			}).Info("get slot")
			epoch := common.SlotToEpoch(int64(slot))
			// generate new strategy at the end of last epoch.
			if int64(slot) < common.EpochEnd(epoch) {
				continue
			}
			if epoch == latestEpoch {
				continue
			}
			latestEpoch = epoch

			{
				nextEpoch := epoch + 1
				cas := 0

				nextDuties, err := attacker.GetEpochDuties(nextEpoch)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
						"epoch": nextEpoch,
					}).Error("failed to get duties")
					latestEpoch = epoch - 1
					continue
				}
				preDuties, err := attacker.GetEpochDuties(epoch - 1)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
						"epoch": epoch - 1,
					}).Error("failed to get duties")
					latestEpoch = epoch - 1
					continue
				}
				curDuties, err := attacker.GetEpochDuties(epoch)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
						"epoch": epoch,
					}).Error("failed to get duties")
					latestEpoch = epoch - 1
					continue
				}
				strategy := types.Strategy{}
				if checkFirstByzSlot(preDuties, params) &&
					checkFirstByzSlot(curDuties, params) &&
					!checkFirstByzSlot(nextDuties, params) {
					cas = 1
				}
				strategy.Uid = uuid.NewString()
				strategy.Slots = GenSlotStrategy(params.FillterHackerDuties(nextDuties), cas)
				if err = attacker.UpdateStrategy(strategy); err != nil {
					log.WithField("error", err).Error("failed to update strategy")
				} else {
					log.WithFields(log.Fields{
						"epoch":    nextEpoch,
						"strategy": strategy,
					}).Info("update strategy successfully")
				}
			}
		}
	}
}

func getLatestHackerSlot(duties []types.ProposerDuty, param types.LibraryParams) int {
	latest, _ := strconv.Atoi(duties[0].Slot)
	for _, duty := range duties {
		idx, _ := strconv.Atoi(duty.ValidatorIndex)
		slot, _ := strconv.Atoi(duty.Slot)
		if !param.IsHackValidator(idx) {
			continue
		}
		if slot > latest {
			latest = slot
		}
	}
	return latest

}

func checkFirstByzSlot(duties []types.ProposerDuty, param types.LibraryParams) bool {
	firstproposerindex, _ := strconv.Atoi(duties[0].ValidatorIndex)
	if !param.IsHackValidator(firstproposerindex) {
		return false
	}
	return true
}

func (o *Instance) Name() string {
	return "staircase"
}

func (o *Instance) Description() string {
	desc_eng := "Staircase attack"
	return desc_eng
}
