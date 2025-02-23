package dp

import (
	"context"
	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	log "github.com/sirupsen/logrus"
	"sync"
)

type impSession struct {
	ctx     context.Context
	client  *deepseek.Client
	Msg     []deepseek.ChatCompletionMessage
	Results []deepseek.Message
	mux     sync.Mutex
	model   string
}

func (s *impSession) Ask(msg string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Msg = append(s.Msg, deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleSystem,
		Content: msg,
	})
	resp, err := s.client.CreateChatCompletion(
		s.ctx,
		&deepseek.ChatCompletionRequest{
			Model:    s.model,
			Messages: s.Msg,
		},
	)

	if err != nil {
		log.WithField("error", err).Error("Create ChatCompletion error")
		return "", err
	}
	toCCM, err := deepseek.MapMessageToChatCompletionMessage(resp.Choices[0].Message)
	if err != nil {
		log.WithField("error", err).Error("Mapping to message failed")
		return "", err
	}
	s.Msg = append(s.Msg, toCCM)
	s.Results = append(s.Results, resp.Choices[0].Message)
	return s.Results[len(s.Results)-1].Content, nil
}
