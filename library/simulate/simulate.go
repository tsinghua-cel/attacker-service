package simulate

import (
	"context"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/types"
	"math/rand"
	"strconv"
	"time"
)

type Instance struct {
}

func (o *Instance) Name() string {
	return "simulate"
}

func (o *Instance) Description() string {
	desc_eng := `Simulate generate random strategy without beacon node.`
	return desc_eng
}

func (o *Instance) Run(ctx context.Context, params types.LibraryParams, feedbacker types.FeedBacker) {
	log.WithField("name", o.Name()).Info("start to run strategy")
	ticker := time.NewTicker(time.Second * 3)
	attacker := params.Attacker
	history := make(map[int]bool)

	secondsPerSlot := 12
	slotsPerEpoch := 32
	simulationStartSlot := int64(0)

	// Simulate the current slot based on the elapsed time since the start of the simulation.
	simulationStartTime := time.Now()
	var getCurSlot = func() int64 {
		elapsed := time.Since(simulationStartTime)
		simulatedSlots := int64(elapsed.Seconds()) / int64(secondsPerSlot)
		return simulationStartSlot + simulatedSlots
	}
	for {
		select {
		case <-ctx.Done():
			log.WithField("name", o.Name()).Info("stop to run strategy")
			return
		case <-ticker.C:
			slot := getCurSlot()
			epoch := slot / int64(slotsPerEpoch)
			nextEpoch := epoch + 1
			log.WithFields(log.Fields{
				"slot":      slot,
				"nextEpoch": nextEpoch,
			}).Info("get slot")

			if _, ok := history[int(nextEpoch)]; ok {
				continue
			}

			duties := GenerateRandomDuty(int(epoch))

			{
				strategy := types.Strategy{}
				strategy.Uid = uuid.NewString()
				strategy.Slots = GenSlotStrategy(params.FilterHackerDuties(duties), nextEpoch)
				strategy.Category = o.Name()
				if err := attacker.UpdateStrategy(strategy); err != nil {
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

func GetRandomUniqueNumbers() []int {
	rand.Seed(time.Now().UnixNano())
	numbers := rand.Perm(256)
	return numbers[:32]
}

func GenerateRandomDuty(epoch int) []types.ProposerDuty {
	startSlot := epoch * 32
	randomSeq := GetRandomUniqueNumbers()
	duties := make([]types.ProposerDuty, 0)
	for i, idx := range randomSeq {
		duty := types.ProposerDuty{
			ValidatorIndex: strconv.Itoa(idx),
			Slot:           strconv.Itoa(startSlot + i),
		}
		duties = append(duties, duty)
	}
	return duties
}
