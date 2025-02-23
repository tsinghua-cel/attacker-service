package dp

import (
	"context"
	"github.com/cohesion-org/deepseek-go"
	"github.com/tsinghua-cel/attacker-service/types"
)

const (
	DeepSeekR1 = "deepseek-r1"
	DeepSeekV3 = "deepseek-v3"
)

type DPAI struct {
	client *deepseek.Client
}

func NewDPAI(key string, url string) types.AIEngine {
	client := deepseek.NewClient(key, url)
	return &DPAI{
		client: client,
	}
}

func (ai *DPAI) NewSession(ctx context.Context, model string) types.AISession {
	s := &impSession{
		model:   model,
		client:  ai.client,
		ctx:     ctx,
		Msg:     []deepseek.ChatCompletionMessage{},
		Results: make([]deepseek.Message, 0),
	}

	return s
}
