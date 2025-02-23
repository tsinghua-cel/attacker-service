package ai

import (
	"github.com/tsinghua-cel/attacker-service/ai/chat"
	"github.com/tsinghua-cel/attacker-service/ai/dp"
	"github.com/tsinghua-cel/attacker-service/types"
)

func GetAI(provider string, key string, url string) types.AIEngine {
	switch provider {
	case "deepseek":
		return dp.NewDPAI(key, url)
	case "openai":
		return chat.NewOpenAI(key, url)
	default:
		return nil
	}

}
