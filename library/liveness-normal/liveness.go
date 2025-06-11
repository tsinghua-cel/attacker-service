package livenessnormal

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
	b     types.ServiceBackend
	param types.LibraryParams
}

func (o *Instance) Name() string {
	return "livenessnormal"
}

func (o *Instance) Description() string {
	// implement attack https://ethresear.ch/t/liveness-attack-in-ethereum-pos-protocol-using-randao-manipulation/22241
	// this strategy is used for honest validators.
	desc_eng := `liveness normal.`
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
	o.b = attacker.GetBackend()
	o.param = params

	t := time.NewTicker(time.Second)
	defer t.Stop()

	history := make(map[int]bool)
	epochDutyCache := lru.New(10)
	var getCacheDuty = func(epoch int64) (duties []types.ProposerDuty) {
		return nil
		//if d, exist := epochDutyCache.Get(epoch); exist {
		//	return d.([]types.ProposerDuty)
		//} else {
		//	return nil
		//}
	}
	var setCacheDuty = func(epoch int64, duties []types.ProposerDuty) {
		epochDutyCache.Add(epoch, duties)
	}
	triggerring := false
	triggeredEpoch := 0 // record the epoch that strategy is triggered.
	for {
		select {
		case <-ctx.Done():
			log.WithField("name", o.Name()).Info("stop to run strategy")
			return
		case <-t.C:
			// 如果当前epoch的第一个slot是attacker, 并且下一个epoch 的第一个slot是attacker,
			// 那么当前epoch 为 epoch 1， 后续一共三个epoch.
			// when strategy triggered, all block is not broadcast.
			//state, err := o.b.GetBeaconState("head")
			//if err != nil {
			//	olog.WithField("error", err).Error("failed to get beacon state")
			//	continue
			//}
			//slot, _ := state.Slot()
			var err error
			slot := common.CurrentSlot()
			epoch := common.SlotToEpoch(int64(slot))
			if history[int(epoch)] == true {
				continue
			}
			nextEpoch := epoch + 1

			var curDuty = getCacheDuty(epoch)
			var nextDuty = getCacheDuty(nextEpoch)
			if curDuty == nil {
				if duty, err := attacker.GetEpochDuties(epoch); err != nil {
					continue
				} else {
					setCacheDuty(epoch, duty)
					curDuty = duty
					//olog.WithFields(log.Fields{
					//	"epoch": epoch,
					//	"duty":  len(duty),
					//}).Info("get epoch duties")
				}
			}
			if nextDuty == nil {
				if duty, err := attacker.GetEpochDuties(nextEpoch); err != nil {
					continue
				} else {
					setCacheDuty(nextEpoch, duty)
					nextDuty = duty
					//
					//olog.WithFields(log.Fields{
					//	"epoch": nextEpoch,
					//	"duty":  len(duty),
					//}).Info("get epoch duties")
				}
			}
			o.dumpDuties(epoch, curDuty)
			o.dumpDuties(nextEpoch, nextDuty)

			for {
				if !triggerring {
					if params.IsHackValidator(toInt(nextDuty[0].ValidatorIndex)) && params.IsHackValidator(toInt(curDuty[0].ValidatorIndex)) &&
						o.attackerInTailN(params.FilterHackerDuties(curDuty), 9) && epoch > 3 {
						triggerring = true
						triggeredEpoch = int(epoch)
						olog.WithFields(log.Fields{
							"current epoch": epoch,
							"next epoch":    epoch + 1,
						}).Info("strategy trigger")
						continue
					}
					history[int(epoch)] = true
					break
				} else {
					offset := epoch - int64(triggeredEpoch) + 1
					if offset == 1 {
						{
							// update current epoch strategy.
							strategy := types.Strategy{}
							strategy.Uid = uuid.NewString()
							strategy.Slots = genStrategyForTrigger(int(epoch), curDuty)
							strategy.Category = o.Name()
							if err = attacker.UpdateStrategy(strategy); err != nil {
								olog.WithField("error", err).Error("failed to update triggering strategy")
							} else {
								olog.WithFields(log.Fields{
									"epoch":    nextEpoch,
									"strategy": strategy,
									"trigger":  triggerring,
									"offset":   1,
								}).Info("update triggering strategy successfully")
							}
						}
						{
							// generate next epoch strategy.
							strategy := types.Strategy{}
							strategy.Uid = uuid.NewString()
							strategy.Slots = genStrategyForTrigger(int(nextEpoch), nextDuty)
							strategy.Category = o.Name()
							if err = attacker.UpdateStrategy(strategy); err != nil {
								olog.WithField("error", err).Error("failed to update triggering strategy")
							} else {
								olog.WithFields(log.Fields{
									"epoch":    nextEpoch,
									"strategy": strategy,
									"trigger":  triggerring,
									"offset":   2,
								}).Info("pre update triggering strategy successfully")
							}
						}
						history[int(epoch)] = true
						break
					} else {
						{
							// generate next epoch strategy.
							strategy := types.Strategy{}
							strategy.Uid = uuid.NewString()
							strategy.Slots = genStrategyForTrigger(int(nextEpoch), nextDuty)
							strategy.Category = o.Name()
							if err = attacker.UpdateStrategy(strategy); err != nil {
								olog.WithField("error", err).Error("failed to update triggering strategy")
							} else {
								olog.WithFields(log.Fields{
									"epoch":    nextEpoch,
									"strategy": strategy,
									"trigger":  triggerring,
									"offset":   3,
								}).Info("update triggering strategy successfully")
							}
						}
						history[int(epoch)] = true
						break
					}
				}
			}
		}
	}
}

func (o *Instance) attackerInTailN(attackDuties []types.ProposerDuty, tailN int) bool {
	latest := attackDuties[len(attackDuties)-1]
	slot := toInt(latest.Slot)
	epoch := common.SlotToEpoch(int64(slot))
	epochEnd := common.EpochEnd(epoch)
	return (int(epochEnd) - tailN) <= slot
}

func (s *Instance) dumpDuties(epoch int64, duties []types.ProposerDuty) {
	for _, duty := range duties {
		log.WithFields(log.Fields{
			"epoch":     epoch,
			"slot":      duty.Slot,
			"validator": duty.ValidatorIndex,
		}).Debug("epoch duty")
	}
}
