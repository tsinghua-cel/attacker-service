package aiattack

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/types"
	"sync"
	"time"
)

type Instance struct {
	ctx        context.Context
	strategies map[string]types.Strategy
	mux        sync.Mutex
	once       sync.Once
}

func (o *Instance) Name() string {
	return "ai"
}

func (o *Instance) Description() string {
	desc_eng := `ai generate strategy.`
	return desc_eng
}

func (o *Instance) Run(ctx context.Context, params types.LibraryParams, feedbacker types.FeedBacker) {
	o.run(ctx, params, feedbacker)
}

func (o *Instance) init() {
	// create ai engine.
	o.strategies = make(map[string]types.Strategy)
	initAgent(o.ctx)
}

func (o *Instance) waitFeedback(ctx context.Context, attacker types.AttackerInc, uid string) (types.FeedBackInfo, error) {
	tm := time.NewTicker(3 * time.Second)
	defer tm.Stop()
	for {
		select {
		case <-ctx.Done():
			return types.FeedBackInfo{}, errors.New("context done")
		case <-tm.C:
			// wait feedback
			feed, err := attacker.GetStrategyFeedback(uid)
			if err == nil {
				return feed, nil
			}
		}
	}
}

func (o *Instance) run(ctx context.Context, params types.LibraryParams, feedbacker types.FeedBacker) {
	o.ctx = ctx
	log.WithField("name", o.Name()).Info("start to run strategy")
	logger := log.WithField("name", o.Name())
	o.once.Do(o.init)
	logger.WithFields(log.Fields{
		"params": params,
	}).Debug("init finished")
	attacker := params.Attacker
	// first
	strategy, err := firstStrategy()
	if err != nil {
		logger.WithError(err).Error("failed to generate first strategy")
		return
	}
	for {
		select {
		case <-ctx.Done():
			logger.Info("stop to run strategy")
			return
		default:
			strategy.Category = o.Name()
			if err = attacker.UpdateStrategy(strategy); err != nil {
				logger.WithField("strategy", strategy).WithError(err).Error("failed to update strategy")
			} else {
				feedback, err := o.waitFeedback(ctx, attacker, strategy.Uid)
				if err != nil {
					logger.WithError(err).Error("failed to get feedback")
				} else {
					feedbackstr := fmt.Sprintf("[\"loss rate of honest validators\": \"%f\",\"loss rate of Byzantine validators\": \"%f\"]",
						feedback.HonestLoseRate, feedback.AttackerLoseRate)
					if strategy, err = newStrategy(feedbackstr); err != nil {
						logger.Error("failed to generate new strategy, stop to run strategy")
						return
					}
				}
			}
		}
	}
}
