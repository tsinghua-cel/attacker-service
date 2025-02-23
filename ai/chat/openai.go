package chat

import (
	"context"
	"github.com/sashabaranov/go-openai"
	"github.com/tsinghua-cel/attacker-service/types"
)

type AI struct {
	client *openai.Client
}

func NewOpenAI(key string, url string) types.AIEngine {
	return &AI{
		client: getClient(key, url),
	}
}

func getClient(key string, url string) *openai.Client {
	cfg := openai.DefaultConfig(key)
	if len(url) != 0 {
		cfg.BaseURL = url
	}

	client := openai.NewClientWithConfig(cfg)
	return client
}

func (ai *AI) NewSession(ctx context.Context, model string) types.AISession {
	s := &impSession{
		model:   model,
		client:  ai.client,
		ctx:     ctx,
		Msg:     []openai.ChatCompletionMessage{},
		Results: make([]openai.ChatCompletionMessage, 0),
	}

	return s
}
