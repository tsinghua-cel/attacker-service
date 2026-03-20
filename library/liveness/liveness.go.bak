package liveness

import (
	"context"
	"github.com/google/uuid"
	"github.com/prysmaticlabs/prysm/v5/cache/lru"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
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
	o.b = attacker.GetBackend()
	o.param = params

	t := time.NewTicker(time.Second)
	defer t.Stop()

	history := make(map[int]bool)
	epochMaskDutiesCache := lru.New(10)
	var getCachedMaskDuties = func(epoch int64) (masked BestMaskDutyInfo) {
		if d, exist := epochMaskDutiesCache.Get(epoch); exist {
			return d.(BestMaskDutyInfo)
		} else {
			return BestMaskDutyInfo{}
		}
	}
	var setCachedMaskDuties = func(epoch int64, masked BestMaskDutyInfo) {
		epochMaskDutiesCache.Add(epoch, masked)
	}

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
			// 当前epoch的第一个slot delay 8s.
			state, err := o.b.GetBeaconState("head")
			if err != nil {
				olog.WithField("error", err).Error("failed to get beacon state")
				continue
			}
			slot, _ := state.Slot()
			epoch := common.SlotToEpoch(int64(slot))
			if history[int(epoch)] == true {
				continue
			}
			nextEpoch := epoch + 1

			var curDuty = getCacheDuty(epoch)
			var nextDuty = getCacheDuty(nextEpoch)
			if curDuty == nil {
				if duty, err := attacker.GetEpochDutiesFromAttack(epoch); err != nil {
					continue
				} else {
					setCacheDuty(epoch, duty)
					curDuty = duty
				}
			}
			if nextDuty == nil {
				if duty, err := attacker.GetEpochDutiesFromAttack(nextEpoch); err != nil {
					continue
				} else {
					setCacheDuty(nextEpoch, duty)
					nextDuty = duty
				}
			}
			o.dumpDuties(epoch, curDuty)
			o.dumpDuties(nextEpoch, nextDuty)
			offset := 0
			// check trigger
			if epoch > 3 && !triggerring && params.IsHackValidator(toInt(nextDuty[0].ValidatorIndex)) && params.IsHackValidator(toInt(curDuty[0].ValidatorIndex)) &&
				o.attackerInTailN(params.FilterHackerDuties(curDuty), 9) {
				triggerring = true
				triggeredEpoch = int(epoch)
			}

			if triggerring {
				offset = int(epoch) - triggeredEpoch + 1 // 1,2,or 3.
			}

			switch offset {
			case 0:
				// check next epoch's first slot is attacker.
				if params.IsHackValidator(toInt(nextDuty[0].ValidatorIndex)) {
					// go to calc best mask duty for next epoch.
					go func() {
						masked, err := o.ComputeBestMask(uint64(slot), nextDuty, AttackerCountCmper{})
						if err != nil {
							olog.WithField("error", err).Error("failed to compute best mask duty")
						} else {
							setCachedMaskDuties(nextEpoch, masked)
						}
					}()
				}

				// generate a simple strategy for nextDuty.
				slotsStrategies := genSimpleStrategy(int(nextEpoch), params.FilterHackerDuties(nextDuty))
				strategy := types.NewStrategy(o.Name(), slotsStrategies, []types.ValidatorStrategy{})
				if err = attacker.UpdateStrategy(strategy); err != nil {
					olog.WithField("error", err).Error("failed to update strategy simple")
				} else {
					olog.WithFields(log.Fields{
						"epoch":    nextEpoch,
						"strategy": strategy,
					}).Debug("update strategy successfully")
				}
				history[int(epoch)] = true

			case 1:
				{
					// go to calc best mask duty for next epoch.
					go func() {
						masked, err := o.ComputeBestMask(uint64(slot), nextDuty, AttackerCountAndFirstAttackerCmper{})
						if err != nil {
							olog.WithField("error", err).Error("failed to compute best mask duty")
						} else {
							setCachedMaskDuties(nextEpoch, masked)
						}
					}()
					{
						curEpochMasked := getCachedMaskDuties(epoch)
						// update current epoch strategy.
						slotsStrategies := genStrategyForTrigger1(int(epoch), params.FilterHackerDuties(curDuty), curEpochMasked)
						strategy := types.NewStrategy(o.Name(), slotsStrategies, []types.ValidatorStrategy{})
						if err = attacker.UpdateStrategy(strategy); err != nil {
							olog.WithField("error", err).Error("failed to update triggering strategy")
						} else {
							olog.WithFields(log.Fields{
								"epoch":    epoch,
								"strategy": strategy,
								"trigger":  triggerring,
								"offset":   1,
							}).Debug("update triggering strategy successfully")
						}
					}
					{
						// generate next epoch strategy without bestMaskDuty.
						strategy := types.Strategy{}
						strategy.Uid = uuid.NewString()
						strategy.Slots = genStrategyForTrigger2(int(nextEpoch), params.FilterHackerDuties(nextDuty), BestMaskDutyInfo{})
						strategy.Category = o.Name()
						if err = attacker.UpdateStrategy(strategy); err != nil {
							olog.WithField("error", err).Error("failed to update triggering strategy")
						} else {
							olog.WithFields(log.Fields{
								"epoch":    nextEpoch,
								"strategy": strategy,
								"trigger":  triggerring,
								"offset":   2,
							}).Debug("pre update triggering strategy successfully")
						}
					}
					history[int(epoch)] = true
				}
			case 2:

				// go to calc best mask duty for next epoch.
				go func() {
					masked, err := o.ComputeBestMask(uint64(slot), nextDuty, AttackerCountAndFirstAttackerCmper{})
					if err != nil {
						olog.WithField("error", err).Error("failed to compute best mask duty")
					} else {
						setCachedMaskDuties(nextEpoch, masked)
					}
				}()
				{
					curEpochMasked := getCachedMaskDuties(epoch)
					// update current epoch strategy.
					strategy := types.Strategy{}
					strategy.Uid = uuid.NewString()
					strategy.Slots = genStrategyForTrigger2(int(epoch), params.FilterHackerDuties(curDuty), curEpochMasked)
					strategy.Category = o.Name()
					if err = attacker.UpdateStrategy(strategy); err != nil {
						olog.WithField("error", err).Error("failed to update triggering strategy")
					} else {
						olog.WithFields(log.Fields{
							"epoch":    nextEpoch,
							"strategy": strategy,
							"trigger":  triggerring,
							"offset":   2,
						}).Debug("update triggering strategy successfully")
					}
				}
				{
					// generate next epoch strategy without bestMaskDuty.
					strategy := types.Strategy{}
					strategy.Uid = uuid.NewString()
					strategy.Slots = genStrategyForTrigger3(int(nextEpoch), params.FilterHackerDuties(nextDuty), BestMaskDutyInfo{})
					strategy.Category = o.Name()
					if err = attacker.UpdateStrategy(strategy); err != nil {
						olog.WithField("error", err).Error("failed to update triggering strategy")
					} else {
						olog.WithFields(log.Fields{
							"epoch":    nextEpoch,
							"strategy": strategy,
							"trigger":  triggerring,
							"offset":   3,
						}).Debug("pre update triggering strategy successfully")
					}
				}
				history[int(epoch)] = true
			case 3:
				{
					// go to calc best mask duty for next epoch.
					go func() {
						masked, err := o.ComputeBestMask(uint64(slot), nextDuty, AttackerCountCmper{})
						if err != nil {
							olog.WithField("error", err).Error("failed to compute best mask duty")
						} else {
							setCachedMaskDuties(nextEpoch, masked)
						}
					}()
					{
						curEpochMasked := getCachedMaskDuties(epoch)
						// update current epoch strategy.
						strategy := types.Strategy{}
						strategy.Uid = uuid.NewString()
						strategy.Slots = genStrategyForTrigger3(int(epoch), params.FilterHackerDuties(curDuty), curEpochMasked)
						strategy.Category = o.Name()
						if err = attacker.UpdateStrategy(strategy); err != nil {
							olog.WithField("error", err).Error("failed to update triggering strategy")
						} else {
							olog.WithFields(log.Fields{
								"epoch":    epoch,
								"strategy": strategy,
								"trigger":  triggerring,
								"offset":   3,
							}).Debug("update triggering strategy successfully")
						}
					}
					{
						// set triggering to false
						triggerring = false
						// generate a simple strategy for nextDuty first.
						slotsStrategies := genSimpleStrategy(int(nextEpoch), params.FilterHackerDuties(nextDuty))
						strategy := types.NewStrategy(o.Name(), slotsStrategies, []types.ValidatorStrategy{})
						if err = attacker.UpdateStrategy(strategy); err != nil {
							olog.WithField("error", err).Error("failed to update strategy simple")
						} else {
							olog.WithFields(log.Fields{
								"epoch":    nextEpoch,
								"strategy": strategy,
							}).Debug("update strategy successfully")
						}

					}
					history[int(epoch)] = true
				}
			}
		}
	}
}

type BestMaskDutyInfo struct {
	FirstIsAttack  bool
	AttackersCount int
	maskedDuties   []types.ProposerDuty // the masked maskedDuties.
	order          []int
	proposers      []primitives.ValidatorIndex
	seed           []byte
}

func (o *Instance) attackerCount(vals []primitives.ValidatorIndex) int {
	var count = 0
	for i := 0; i < len(vals); i++ {
		if o.param.IsHackValidator(int(vals[i])) {
			count++
		}
	}
	return count
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
