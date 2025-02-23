package ai

import (
	"context"
	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

func TestMultiChat(t *testing.T) {
	key := os.Getenv("OPENAI_API_KEY")
	url := os.Getenv("OPENAI_BASE_URL")
	client := deepseek.NewClient(key, url)
	ctx := context.Background()

	messages := []deepseek.ChatCompletionMessage{{
		Role:    constants.ChatMessageRoleUser,
		Content: "Who is the president of the United States? One word response only.",
	}}

	// Round 1: First API call
	response1, err := client.CreateChatCompletion(ctx, &deepseek.ChatCompletionRequest{
		Model:    "deepseek-r1",
		Messages: messages,
	})
	if err != nil {
		log.Fatalf("Round 1 failed: %v", err)
	}

	response1Message, err := deepseek.MapMessageToChatCompletionMessage(response1.Choices[0].Message)
	if err != nil {
		log.Fatalf("Mapping to message failed: %v", err)
	}
	messages = append(messages, response1Message)

	log.Printf("The messages after response 1 are: %v", messages)
	// Round 2: Second API call
	messages = append(messages, deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleUser,
		Content: "Who was the one in the previous term.",
	})

	response2, err := client.CreateChatCompletion(ctx, &deepseek.ChatCompletionRequest{
		Model:    "deepseek-r1",
		Messages: messages,
	})
	if err != nil {
		log.Fatalf("Round 2 failed: %v", err)
	}

	response2Message, err := deepseek.MapMessageToChatCompletionMessage(response2.Choices[0].Message)
	if err != nil {
		log.Fatalf("Mapping to message failed: %v", err)
	}
	messages = append(messages, response2Message)
	log.Printf("The messages after response 1 are: %v", messages)

}
