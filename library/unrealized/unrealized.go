package unrealized

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
	return "unrealized"
}

func (o *Instance) Description() string {
	desc_eng := `Unrealized justification reorg attack`
	return desc_eng
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
			epoch := common.SlotToEpoch(int64(slot))
			if epoch == latestEpoch {
				continue
			}
			if (int64(slot) + 5) < common.EpochEnd(epoch) {
				continue
			}
			latestEpoch = epoch
			// get next epoch duties
			duties, err := attacker.GetEpochDuties(epoch + 1)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"epoch": epoch + 1,
				}).Error("failed to get duties")
				latestEpoch = epoch - 1
				continue
			}
			if hackDuties, happen := CheckDuties(params, duties); happen {
				strategy := types.Strategy{}
				strategy.Uid = uuid.NewString()
				strategy.Slots = GenSlotStrategy(hackDuties)
				if err = attacker.UpdateStrategy(strategy); err != nil {
					log.WithField("error", err).Error("failed to update strategy")
				} else {
					log.WithFields(log.Fields{
						"epoch":    epoch + 1,
						"strategy": strategy,
					}).Info("update strategy successfully")
				}
			}
		}
	}
}
