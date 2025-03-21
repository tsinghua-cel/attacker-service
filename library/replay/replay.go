package replay

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/dbmodel"
	"github.com/tsinghua-cel/attacker-service/types"
	"time"
)

type Instance struct{}

func (o *Instance) Run(ctx context.Context, params types.LibraryParams, feedbacker types.FeedBacker) {
	log.WithField("name", o.Name()).Info("start to run strategy")
	if params.Extend == nil || params.Extend["replay"] == "" {
		log.WithField("name", o.Name()).Error("replay project id is empty")
		return
	}
	var replayProject = params.Extend["replay"]
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
			if int64(slot) < common.EpochEnd(epoch) {
				continue
			}
			latestEpoch = epoch
			next := epoch + 1
			// get next epoch duties
			originStrategy := dbmodel.GetStrategyByProjectAndEpoch(replayProject.(string), next)
			if originStrategy != nil {
				strategy := types.Strategy{}
				if err := json.Unmarshal([]byte(originStrategy.Content), &strategy); err != nil {
					log.WithField("uuid", originStrategy.UUID).WithField("error", err).Error("failed to unmarshal strategy")
					continue
				}
				strategy.Uid = uuid.NewString()
				log.WithField("epoch", next).WithField("project", replayProject).WithField("strategy", strategy).Info("replay strategy")
				if err := attacker.UpdateStrategy(strategy); err != nil {
					log.WithField("error", err).Error("failed to update strategy")
				}
			} else {
				log.WithField("epoch", next).WithField("project", replayProject).Error("no strategy found, skip it")
			}
		}
	}
}

func (o *Instance) Name() string {
	return "replay"
}

func (o *Instance) Description() string {
	desc_eng := `Replay all strategies that generate at a special project id.`
	return desc_eng
}
