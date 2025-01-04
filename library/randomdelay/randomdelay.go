package randomdelay

import (
	"context"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"sync"
	"time"
)

type Instance struct {
	strategies map[string]types.Strategy
	mux        sync.Mutex
	once       sync.Once
}

func (o *Instance) Name() string {
	return "random"
}

func (o *Instance) Description() string {
	desc_eng := `generate random actions.`
	return desc_eng
}

func (o *Instance) init() {
	o.strategies = make(map[string]types.Strategy)
}

func (o *Instance) Run(ctx context.Context, params types.LibraryParams, feedbacker types.FeedBacker) {
	log.WithField("name", o.Name()).Info("start to run strategy")
	o.once.Do(o.init)

	feedbackCh := make(chan types.FeedBack, 100)
	updateFeedBack := func() {
		for {
			select {
			case info, ok := <-feedbackCh:
				if !ok {
					return
				}
				o.mux.Lock()
				if s, exist := o.strategies[info.Uid]; exist {
					log.WithFields(log.Fields{
						"uid":  info.Uid,
						"info": info.Info,
					}).Info("get feedback")
					if info.Info.ReorgCount > 0 || info.Info.ImpactValidatorCount > 0 {
						log.WithFields(log.Fields{
							"uid":      info.Uid,
							"strategy": s.String(),
						}).Info("feedback have good impact, please save it")
					}
					// todo: update strategy with feedback.
				}
				o.mux.Unlock()
			}
		}
	}
	if feedbacker != nil {
		go updateFeedBack()
	}

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
			log.WithFields(log.Fields{
				"slot": slot,
				"name": o.Name(),
			}).Info("check strategy")
			epoch := common.SlotToEpoch(int64(slot))
			if epoch == latestEpoch {
				continue
			}
			if int64(slot) < common.EpochEnd(epoch) {
				// only update strategy at the end of current epoch.
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
					log.WithField("strategy", strategy).WithError(err).Error("failed to update strategy")
				} else {
					log.WithFields(log.Fields{
						"epoch":    epoch + 1,
						"strategy": strategy.Uid,
					}).Info("update strategy successfully")
					if feedbacker != nil {
						o.mux.Lock()
						o.strategies[strategy.Uid] = strategy
						o.mux.Unlock()
						feedbacker.WaitFeedback(strategy.Uid, feedbackCh)
					}
				}
			}
		}
	}
}
