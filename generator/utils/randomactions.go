package utils

import (
	"fmt"
	"github.com/tsinghua-cel/attacker-service/common"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type ActionPoint string

const (
	AttestBeforeBroadCast     ActionPoint = "AttestBeforeBroadCast"
	AttestAfterBroadCast      ActionPoint = "AttestAfterBroadCast"
	AttestBeforeSign          ActionPoint = "AttestBeforeSign"
	AttestAfterSign           ActionPoint = "AttestAfterSign"
	AttestBeforePropose       ActionPoint = "AttestBeforePropose"
	AttestAfterPropose        ActionPoint = "AttestAfterPropose"
	BlockDelayForReceiveBlock ActionPoint = "BlockDelayForReceiveBlock"
	BlockBeforeBroadCast      ActionPoint = "BlockBeforeBroadCast"
	BlockAfterBroadCast       ActionPoint = "BlockAfterBroadCast"
	BlockBeforeSign           ActionPoint = "BlockBeforeSign"
	BlockAfterSign            ActionPoint = "BlockAfterSign"
	BlockBeforePropose        ActionPoint = "BlockBeforePropose"
	BlockAfterPropose         ActionPoint = "BlockAfterPropose"
	BlockGetNewParentRoot     ActionPoint = "BlockGetNewParentRoot"
)

func (a ActionPoint) SpecialActions() []string {
	switch a {
	case "BlockGetNewParentRoot":
		return []string{"modifyParentRoot:[slot]"}
	case "AttestBeforeSign":
		return []string{"modifyAttestSource:[slot]", "modifyAttestTarget:[slot]", "modifyAttestHead:[slot]"}
	default:
		return []string{}
	}
}

func (a ActionPoint) FullActions() []string {
	normalActions := []string{
		"return",
		"delayWithSecond:[time]",
	}
	normalActions = append(normalActions, a.SpecialActions()...)
	return normalActions
}

func (a ActionPoint) RandomAction(slot int) string {
	actions := a.FullActions()
	action := actions[rand.Intn(len(actions))]
	if strings.Contains(action, "[slot]") {
		rparam := rand.Intn(30)
		if slot > rparam {
			rparam = slot - rparam
		}
		action = strings.Replace(action, "[slot]", strconv.FormatInt(int64(rparam), 10), -1)
	} else {
		base := common.GetChainBaseInfo()
		rparam := (rand.Intn(10) + 3) * base.SecondsPerSlot
		action = strings.Replace(action, "[time]", strconv.Itoa(rparam), -1)
	}
	return action
}

var (
	beforePoint = []ActionPoint{
		AttestBeforeBroadCast,
		AttestBeforeSign,
		AttestBeforePropose,
		BlockDelayForReceiveBlock,
		BlockBeforeBroadCast,
		BlockBeforeSign,
		BlockBeforePropose,
		BlockGetNewParentRoot,
	}
)

type RSet struct {
	items []ActionPoint
}

func (s *RSet) Add(item []ActionPoint) {
	s.items = append(s.items, item...)
}

func (s *RSet) PopRandom() (ActionPoint, error) {
	if len(s.items) == 0 {
		return ActionPoint(""), fmt.Errorf("set is empty")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := r.Intn(len(s.items))
	item := s.items[index]
	s.items = append(s.items[:index], s.items[index+1:]...)

	return item, nil
}

func GetRandomActions(slot int, maxCount int) map[string]string {
	actions := make(map[string]string)
	rset := &RSet{}
	rset.Add(beforePoint)
	for i := 0; i < maxCount; i++ {
		point, err := rset.PopRandom()
		if err != nil {
			break
		}
		actions[string(point)] = point.RandomAction(slot)
	}
	return actions
}
