package five

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
	return "five"
}

func (o *Instance) Description() string {
	desc_eng := `	Check the order of malicious nodes in the next epoch one epoch in advance;
	Two malicious nodes are interspersed with an honest node's block, so that the parent of the second malicious node's block points to the slot of the previous malicious node.
	Delay strategy: the block of the first malicious node broadcasts a delay of 1 slot.`
	return desc_eng
}

func (o *Instance) Run(ctx context.Context, params types.LibraryParams, feedbacker types.FeedBacker) {
	log.WithField("name", o.Name()).Info("start to run strategy")
	var latestEpoch int64
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
			if int64(slot) < common.EpochEnd(epoch) {
				//continue
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
