package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/tsinghua-cel/attacker-service/ai/dp"
	"github.com/tsinghua-cel/attacker-service/types"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestDeepSeekEth(t *testing.T) {
	prompts, err := os.ReadFile("prompt.txt")
	if err != nil {
		t.Fatalf("read prompt.txt failed: %v", err)
	}
	key := os.Getenv("OPENAI_API_KEY")
	url := os.Getenv("OPENAI_BASE_URL")
	ai := GetAI("openai", key, url)
	if ai == nil {
		t.Error("GetAI() failed")
	}
	session := ai.NewSession(context.TODO(), dp.DeepSeekR1)
	content, err := session.Ask(string(prompts))
	if err != nil {
		t.Error("session.Ask() failed")
	} else {
		fmt.Printf("first response: %s\n", content)
	}
	for {
		time.Sleep(3 * time.Second)
		feedback := fmt.Sprintf("[\"loss rate of honest validators\": \"%s\",\"loss rate of Byzantine validators\": \"%s\"]",
			fmt.Sprintf("%f", rand.Float64()), fmt.Sprintf("%f", rand.Float64()))
		// get input from stdin
		content, err := session.Ask(feedback)
		if err != nil {
			t.Error("session.Ask() failed")
		} else {
			fmt.Printf("response: %s\n", content)
		}

	}
}
func TestSession(t *testing.T) {
	key := os.Getenv("OPENAI_API_KEY")
	url := os.Getenv("OPENAI_BASE_URL")
	ai := GetAI("openai", key, url)

	session := ai.NewSession(context.TODO(), openai.GPT4oMini)
	for {
		content, err := session.Ask("what's your name")
		if err != nil {
			t.Error("session.Ask() failed")
		} else {
			fmt.Printf("response: %s\n", content)
		}
		content, err = session.Ask("your name is Kele")
		if err != nil {
			t.Error("session.Ask() failed")
		} else {
			fmt.Printf("response: %s\n", content)
		}
		time.Sleep(10 * time.Second)

	}
}

func TestGenerateStrategy(t *testing.T) {
	prompt := `
You are a penetration test expert who knows a lot about the principles of blockchain attacks 
and has extensive knowledge of the Ethereum PoS consensus algorithm. 
I need you to help me design a penetration test plan to find potential problems with the Ethereum PoS client prysm.
The test plan format is shown in json. 
The slot is set to a specific slot in the current epoch, the level represents the priority of the plan, 
and when there are multiple test plans under the same slot, the plan with the highest priority is executed.
Actions represents a set of penetration lists, where point is the penetration injection point 
and action is the attack function to be executed.

Test plan format:
[{
	"slot": "1",
	"level": 1,
	"actions":{
		"point1": "action1",
		"point2": "action2",
		...

	}
}]

All your response should be in this format only.

The penetration injection points are divided into two lists. 
The first one is related to the production block, and the injection point list is [
	"BlockDelayForReceiveBlock",
	"BlockGetNewParentRoot",
	"BlockBeforeBroadCast",
	"BlockAfterBroadCast",
	"BlockBeforeSign",
	"BlockAfterSign",
	"BlockBeforePropose",
	"BlockAfterPropose",
], and the second one is related to voting. The injection point list is [
	"AttestBeforeBroadCast",
	"AttestAfterBroadCast",
	"AttestBeforeSign",
	"AttestAfterSign",
	"AttestBeforePropose",
	"AttestAfterPropose",
]. 
All attack functions are divided into three groups. 
The first group is the functions that can be used by all injection points, including ["null", "return", 
"continue", "abort", "exit", "delayWithSecond:", "delayToNextSlot","delayToAfterNextSlot:", "delayToNextNEpochStart:",
"delayToNextNEpochEnd:", "delayToNextNEpochHalf:", "delayToEpochEnd", "delayHalfEpoch"]; 
the second group is the functions specific to block production, "packPooledAttest" specific to "BlockBeforeBroadCast" 
and "modifyParentRoot:" specific to "BlockDelayForReceiveBlock"; 
the third group is the functions specific to voting production, "addAttestToPool" specific to "AttestAfterSign". 
"addAttestToPool" and "packPooledAttest" are often used in combination. The : in the list means that an integer parameter can be added after :.
`
	key := os.Getenv("OPENAI_API_KEY")
	url := os.Getenv("OPENAI_BASE_URL")
	ai := GetAI("openai", key, url)
	session := ai.NewSession(context.TODO(), openai.GPT4oMini)
	beginSlot := 100
	slotsPerEpoch := 32
	first, err := session.Ask(prompt)
	if err != nil {
		t.Error("session.Ask() failed")
	} else {
		fmt.Printf("first response: %s\n", first)
	}
	for i := 0; i < 1; i++ {
		//content, err := session.Ask(fmt.Sprintf("produce several penetration test plans for me and return the plans in json format only.", beginSlot, beginSlot+slotsPerEpoch))
		slots := fmt.Sprintf("%d,%d,%d,%d,%d", rand.Intn(slotsPerEpoch)+beginSlot,
			rand.Intn(slotsPerEpoch)+beginSlot,
			rand.Intn(slotsPerEpoch)+beginSlot,
			rand.Intn(slotsPerEpoch)+beginSlot,
			rand.Intn(slotsPerEpoch)+beginSlot)
		content, err := session.Ask(fmt.Sprintf("produce several penetration test plans for me and slot should between [%s].", slots))
		if err != nil {
			t.Error("session.Ask() failed")
		} else {
			//fmt.Printf("ai: %s\n", content)
		}
		content = strings.Replace(content, "```json", "", -1)
		content = strings.Replace(content, "```", "", -1)
		s := types.Strategy{}
		if err := json.Unmarshal([]byte(content), &s.Slots); err != nil {
			t.Error("json.Unmarshal() failed", err)
			fmt.Println(content)
		} else {
			fmt.Printf("strategy: %v\n", s)
		}
		//time.Sleep(10 * time.Second)
		beginSlot += slotsPerEpoch
	}
}
