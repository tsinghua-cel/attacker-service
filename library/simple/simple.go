package simple

import (
	"context"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"time"
)

type Instance struct {
}

func (o *Instance) Name() string {
	return "simple"
}

func (o *Instance) Description() string {
	desc_eng := `Simple random generate some strategy.`
	return desc_eng
}

func (o *Instance) Run(ctx context.Context, params types.LibraryParams, feedbacker types.FeedBacker) {
	log.WithField("name", o.Name()).Info("start to run strategy")
	ticker := time.NewTicker(time.Second * 3)
	attacker := params.Attacker
	history := make(map[int]bool)
	for {
		select {
		case <-ctx.Done():
			log.WithField("name", o.Name()).Info("stop to run strategy")
			return
		case <-ticker.C:
			slot := attacker.GetCurSlot()
			epoch := common.SlotToEpoch(int64(slot))
			nextEpoch := epoch + 1
			log.WithFields(log.Fields{
				"slot":      slot,
				"nextEpoch": nextEpoch,
			}).Info("get slot")

			if _, ok := history[int(nextEpoch)]; ok {
				continue
			}

			{
				duties, err := attacker.GetEpochDuties(nextEpoch)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
						"epoch": nextEpoch,
					}).Error("failed to get duties")
					continue
				}
				strategy := types.Strategy{}
				strategy.Uid = uuid.NewString()
				strategy.Slots = GenSlotStrategy(params.FillterHackerDuties(duties), nextEpoch)
				strategy.Category = o.Name()
				if err = attacker.UpdateStrategy(strategy); err != nil {
					log.WithField("error", err).Error("failed to update strategy")
				} else {
					log.WithFields(log.Fields{
						"epoch":    nextEpoch,
						"strategy": strategy,
					}).Info("update strategy successfully")
					history[int(nextEpoch)] = true
				}
			}
		}
	}
}
