package confuse

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/types"
	"math/rand"
	"strconv"
)

var (
	epochPerStage = int64(8) // every 8 epoch is a stage, to change strategy.
	stageCache, _ = lru.New(100)
)

type stageInfo struct {
	duties  []types.ProposerDuty
	endSlot int64
}

func BlockStrategy(idx, cur, end int, actions map[string]string) {
	if cur != end {
		// not propose block.
		// add attest to pool, not broadcast.
		actions["BlockBeforeSign"] = "return"
		actions["AttestAfterSign"] = fmt.Sprintf("addAttestToPool")
		actions["AttestBeforeBroadCast"] = fmt.Sprintf("return")
	}
	if cur == end {
		// propose block, and pack all pooled attest.
		// modify parent to last epoch endslot.
		actions["BlockBeforeSign"] = "packPooledAttest"
		lastEpoch := common.SlotToEpoch(int64(end)) - 1
		if data, exist := stageCache.Get(lastEpoch); exist {
			info := data.(stageInfo)
			actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%d", info.endSlot)
		}
	}
}

func BlockStrategy2(idx, cur, end int, actions map[string]string) {
	// all block broadcast delay to end slot.
	// a-b-c-d-e delay to e + 1, e + 2, e + 3, e + 4, e + 0.

	if cur != end {
		// block broadcast delay to end slot.
		delay := (end-cur)*common.GetChainBaseInfo().SecondsPerSlot + idx + 1
		actions["BlockBeforeBroadCast"] = fmt.Sprintf("delayWithSecond:%d", delay)
		// attest broadcast delay to end slot.
		actions["AttestBeforeBroadCast"] = fmt.Sprintf("delayWithSecond:%d", delay)
	}
}

func BlockStrategy3(idx, cur, lastSlot int, actions map[string]string) {
	// modify parent root to lastSlot.
	actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%d", lastSlot)
	// don't broadcast attest.
	actions["AttestBeforeBroadCast"] = fmt.Sprintf("return")
}

func BlockStrategy4(idx, cur, end int, actions map[string]string) {
	// random generate a new strategy.
	// block and attest broadcast delay random time.
	// block modify parent root to last epoch end slot.
	blockPoints := []string{"BlockBeforeBroadCast", "BlockBeforeSign"}
	attestPoints := []string{"AttestBeforeBroadCast", "AttestBeforeSign"}
	blockpoint := blockPoints[rand.Intn(len(blockPoints))]
	attestpoint := attestPoints[rand.Intn(len(attestPoints))]
	actions[blockpoint] = fmt.Sprintf("delayWithSecond:%d", rand.Intn(80)+10)
	actions[attestpoint] = fmt.Sprintf("delayWithSecond:%d", rand.Intn(80)+10)
	lastEpoch := common.SlotToEpoch(int64(end)) - 1
	if data, exist := stageCache.Get(lastEpoch); exist {
		info := data.(stageInfo)
		actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%d", rand.Intn(int(info.endSlot))+1)
	}
}

func BlockStrategy5(idx, cur, end int, actions map[string]string) {
	// random generate a new strategy.
	// block and attest broadcast delay random time.
	// block modify parent root to last epoch end slot.
	cnt := rand.Intn(3) + 1
	blockPoints := types.BlockPointSet
	attestPoints := types.AttestPointSet
	used := make(map[string]bool)
	for i := 0; i < cnt; i++ {
		blockpoint := blockPoints[rand.Intn(len(blockPoints))]
		attestpoint := attestPoints[rand.Intn(len(attestPoints))]
		if used[blockpoint] || used[attestpoint] {
			i--
			continue
		}

		actions[blockpoint] = fmt.Sprintf("delayWithSecond:%d", rand.Intn(80)+10)
		actions[attestpoint] = fmt.Sprintf("delayWithSecond:%d", rand.Intn(80)+10)
		used[blockpoint] = true
		used[attestpoint] = true
	}
	lastEpoch := common.SlotToEpoch(int64(end)) - 1
	if data, exist := stageCache.Get(lastEpoch); exist {
		info := data.(stageInfo)
		actions["BlockGetNewParentRoot"] = fmt.Sprintf("modifyParentRoot:%d",
			rand.Intn(cur-int(info.endSlot))+int(info.endSlot)-common.GetChainBaseInfo().SlotsPerEpoch)
	}
}

func GenSlotStrategy(allHacks []types.ProposerDuty) []types.SlotStrategy {
	if len(allHacks) == 0 {
		return nil
	}
	strategys := make([]types.SlotStrategy, 0)
	duties := make([]types.ProposerDuty, 0)
	latestDuty := allHacks[len(allHacks)-1]
	endSlot, _ := strconv.ParseInt(latestDuty.Slot, 10, 64)
	epoch := common.SlotToEpoch(endSlot)
	lastEpochInfo, haveLast := stageCache.Get(epoch - 1)
	for i, duty := range allHacks {
		duties = append(duties, duty)
		slot, _ := strconv.ParseInt(duty.Slot, 10, 64)
		strategy := types.SlotStrategy{
			Slot:    duty.Slot,
			Level:   1,
			Actions: make(map[string]string),
		}
		var lastSlot int64 = slot - 1 // normally last slot.
		if i == 0 && haveLast {
			lastSlot = lastEpochInfo.(stageInfo).endSlot
		} else if i > 0 {
			lastSlot, _ = strconv.ParseInt(allHacks[i-1].Slot, 10, 64)
		}

		stage := rand.Intn(4) + 1
		//stage := epoch / epochPerStage
		switch stage {
		case 1:
			BlockStrategy(i, int(slot), int(endSlot), strategy.Actions)
		case 2:
			BlockStrategy2(i, int(slot), int(endSlot), strategy.Actions)
		case 3:
			BlockStrategy3(i, int(slot), int(lastSlot), strategy.Actions)
		case 4:
			BlockStrategy4(i, int(slot), int(endSlot), strategy.Actions)
		case 5:
			BlockStrategy5(i, int(slot), int(endSlot), strategy.Actions)
		default:
			BlockStrategy(i, int(slot), int(endSlot), strategy.Actions)
		}

		strategys = append(strategys, strategy)
	}
	stageCache.Add(epoch, stageInfo{duties: duties, endSlot: endSlot})

	return strategys

}
