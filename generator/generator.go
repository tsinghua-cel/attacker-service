package generator

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/generator/utils"
	"github.com/tsinghua-cel/attacker-service/library"
	"github.com/tsinghua-cel/attacker-service/types"
	"math/rand"
	"strings"
	"time"
)

type Generator struct {
	todoStrategies []library.Strategy
	param          types.StrategyGeneratorParam
	attacker       types.AttackerInc
	logger         *log.Entry
}

func NewGenerator(backend types.ServiceBackend, param types.StrategyGeneratorParam) *Generator {
	library.Init()
	strategies := strings.Split(param.Strategy, ",")
	added := make(map[string]bool)
	for _, s := range strategies {
		if s == "all" || library.GetStrategy(s) != nil {
			added[s] = true
		} else {
			log.WithField("strategy", s).Error("skip not founded strategy")
		}
	}
	allStrategy := library.GetAllStrategies()
	addedStrategies := make([]library.Strategy, 0)
	for k, v := range allStrategy {
		if added["all"] || added[k] {
			log.WithField("strategy", k).Info("add strategy")
			addedStrategies = append(addedStrategies, v)
		}
	}
	return &Generator{
		todoStrategies: addedStrategies,
		param:          param,
		attacker:       utils.WrapToAttacker(backend),
		logger:         log.WithField("module", "generator"),
	}
}

func (g *Generator) Start() error {
	go g.runGenerator()
	return nil
}

func (g *Generator) runGenerator() {
	g.logger.Info("generator start")
	defer g.logger.Info("generator exit")
	filtered := make([]library.Strategy, len(g.todoStrategies))
	copy(filtered, g.todoStrategies)

	if len(filtered) == 0 {
		g.logger.Error("no strategy to run")
		return
	}

	if len(filtered) > 1 {
		for {
			randIdx := rand.Intn(len(filtered))
			ctx, cancle := context.WithTimeout(context.Background(), time.Duration(g.param.DurationPerStrategy)*time.Minute)
			strategy := filtered[randIdx]
			strategy.Run(ctx, types.LibraryParams{
				Attacker:          g.attacker,
				MaxValidatorIndex: g.param.MaxMaliciousIdx,
				MinValidatorIndex: g.param.MinMaliciousIdx,
			}, nil)
			cancle()
		}
	} else {
		strategy := filtered[0]
		strategy.Run(context.Background(), types.LibraryParams{
			Attacker:          g.attacker,
			MaxValidatorIndex: g.param.MaxMaliciousIdx,
			MinValidatorIndex: g.param.MinMaliciousIdx,
		}, nil)
	}
}
