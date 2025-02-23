package aiattack

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/ai"
	"github.com/tsinghua-cel/attacker-service/types"
	"os"
	"regexp"
	"strings"
	"time"
)

//go:embed prompt.txt
var prompt string

var (
	agent types.AISession
)

type Param struct {
	Key   string `json:"key"`
	Url   string `json:"url"`
	Model string `json:"model"`
}

func initAgent(ctx context.Context) {
	var param Param
	param.Url = os.Getenv("OPENAI_BASE_URL")
	param.Key = os.Getenv("OPENAI_API_KEY")
	param.Model = os.Getenv("LLM_MODEL")
	log.WithFields(log.Fields{
		"param": param,
		//"prompt": prompt,
	}).Info("initAgent")

	engine := ai.GetAI("openai", param.Key, param.Url)
	if engine == nil {
		log.Error("GetAI() failed")
		return
	} else {
		agent = engine.NewSession(ctx, param.Model)
	}
}

func getJson(content string) string {
	content = strings.Replace(content, "```json", "", -1)
	content = strings.Replace(content, "```", "", -1)
	content = strings.Replace(content, "\n", "", -1)
	content = strings.TrimSpace(content)
	re := regexp.MustCompile("\\[.*\\]")
	jsonStr := re.FindString(content)
	return jsonStr
}

func firstStrategy() (types.Strategy, error) {
	if agent == nil {
		return types.Strategy{}, errors.New("agent is nil")
	}
	for i := 0; i < 10; i++ {
		if i > 0 {
			time.Sleep(3 * time.Second)
		}
		content, err := agent.Ask(prompt)
		if err != nil {
			log.WithError(err).Error("agent.Ask() failed, retry")
			continue
		}
		jsonStr := getJson(content)
		if len(jsonStr) == 0 {
			log.WithField("content", content).Error("getJson() failed")
		} else {
			log.WithField("jsonStr", jsonStr).Info("strategy content")
		}
		var s types.Strategy
		if err = json.Unmarshal([]byte(jsonStr), &s.Slots); err != nil {
			log.WithField("jsonstr", jsonStr).WithError(err).Error("json.Unmarshal() failed, retry")
			continue
		}
		s.Uid = uuid.NewString()
		log.WithField("strategy", s).Debug("first strategy success")
		return s, nil

	}
	return types.Strategy{}, errors.New("first strategy failed")
}

func newStrategy(feedback string) (types.Strategy, error) {
	if agent == nil {
		return types.Strategy{}, errors.New("agent is nil")
	}
	for i := 0; i < 10; i++ {
		if i > 0 {
			time.Sleep(3 * time.Second)
		}
		content, err := agent.Ask(feedback)
		if err != nil {
			log.WithError(err).Error("agent.Ask() failed retry")
			continue
		}
		jsonStr := getJson(content)
		if len(jsonStr) == 0 {
			log.WithField("content", content).Error("getJson() failed")
		} else {
			log.WithField("jsonStr", jsonStr).Info("strategy content")
		}
		var s types.Strategy
		if err = json.Unmarshal([]byte(jsonStr), &s.Slots); err != nil {
			log.WithField("jsonstr", jsonStr).WithError(err).Error("json.Unmarshal() failed retry")
			continue
		}
		s.Uid = uuid.NewString()
		log.WithField("strategy", s).Debug("new strategy success")
		return s, nil
	}
	return types.Strategy{}, errors.New("new strategy failed")
}
