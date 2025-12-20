package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/helpers"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/signing"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	"github.com/OffchainLabs/prysm/v7/crypto/bls"
	"github.com/OffchainLabs/prysm/v7/crypto/hash"
	"github.com/OffchainLabs/prysm/v7/encoding/bytesutil"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/disguisedRandao"
	"github.com/tsinghua-cel/attacker-service/types"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	beaconUrl     = flag.String("beacon-url", "", "Beacon URL")
	stateFile     = flag.String("state", "", "State file")
	validatorList = flag.String("validator-list", "", "Validator list file path")
	slotToTest    = flag.String("slot", "161", "Slot to test")
	testCase      = flag.Int("case", 1, "Test case number(1,2,3")
)

var (
	allIndices = disguisedRandao.GenValidatorIndices(0, 255)
)

func main() {
	flag.Parse()
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})
	// get all validators from json.
	validators, err := getValidatorListFromFile(*validatorList)
	if err != nil {
		log.WithFields(log.Fields{
			"filePath": *validatorList,
			"err":      err,
		}).Fatal("failed to get validator list from file")
		return
	}
	var state *spec.VersionedBeaconState
	if *stateFile != "" {
		data, err := os.ReadFile(*stateFile)
		if err != nil {
			log.WithFields(log.Fields{
				"filePath": *stateFile,
				"err":      err,
			}).Warning("failed to read state file")
			return
		}
		var localstate spec.VersionedBeaconState
		if err = json.Unmarshal(data, &localstate); err != nil {
			log.WithFields(log.Fields{
				"filePath": *stateFile,
				"err":      err,
			}).Fatal("failed to unmarshal state from file")
			return
		}
		state = &localstate
	}
	if state == nil {
		beaconClient := beaconapi.NewBeaconGwClient(*beaconUrl)
		chainstate, err := beaconClient.GetBeaconState(*slotToTest)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal("failed to get beacon state")
			return
		}
		state = chainstate
	}
	if *stateFile == "" {

		statedata, err := json.Marshal(state)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal("failed to marshal beacon state")
			return
		}
		os.WriteFile("state.json", statedata, 0644)
	}
	common.InitSlotTool(12, 32, GetGenesisTime(state))

	chainValidators, err := state.Validators()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("failed to get validators from beacon state")
		return
	}

	validatorSortedIndex := make([]ValidatorInfo, len(chainValidators))
	for index, v := range chainValidators {
		pubkey := v.PublicKey.String()
		if info, ok := validators[pubkey]; ok {
			validatorSortedIndex[index] = info
		} else {
			log.WithFields(log.Fields{
				"pubkey": pubkey,
			}).Warn("validator not found in input list")
		}
	}
	mostate, err := disguisedRandao.InitMoState(*state)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to init mo state")
		return
	}
	//mostate.Dump()
	curSlot, _ := state.Slot()
	epoch := curSlot / 32

	proposerDuty := GenerateRandomDuty(validatorSortedIndex, int(epoch))

	type TestSetting struct {
		attackCount int
		testCount   int
	}
	var testSet = []TestSetting{
		{10, 20},
		{15, 20},
		{20, 20},
	}
	for _, setting := range testSet {
		attackDutiesCount := setting.attackCount
		testCount := setting.testCount
		fullOrder := GetAllBinarySequences(attackDutiesCount)
		targetEpoch := primitives.Epoch(epoch + 2)
		attackDuties := RandomAttackerDuties(rand.New(rand.NewSource(time.Now().UnixNano())), proposerDuty, attackDutiesCount)

		cState := mostate.Reset()
		seed, _ := disguisedRandao.Seed(cState, targetEpoch, disguisedRandao.DomainBeaconProposer)
		t1 := time.Now()
		allRandaoReveal, _ := GetAllRandaoReveal(cState, int64(epoch), attackDuties, validatorSortedIndex)
		t2 := time.Now()
		log.WithFields(log.Fields{
			"cost":         t2.Sub(t1).String(),
			"attack_count": attackDutiesCount,
		}).Info("get all randao reveal for attacker duties")

		//log.Println("full order count ", len(fullOrder))
		switch *testCase {
		case 1:
			{
				// run ComputeBestMaskDutyOneOrderSync testCount times and record durations
				durations := make([]time.Duration, 0, testCount)
				successes := 0
				for i := 0; i < testCount; i++ {
					tStart := time.Now()
					_, err := ComputeBestMaskDutyOneOrderSync(allRandaoReveal, cState.Reset(), uint64(curSlot), int64(epoch), attackDuties, validatorSortedIndex, fullOrder[0])
					elapsed := time.Since(tStart)
					if err != nil {
						log.WithFields(log.Fields{"err": err, "iteration": i}).Error("ComputeBestMaskDutyOneOrderSync failed")
						// continue to next iteration, don't include failed run in averages
						continue
					}
					durations = append(durations, elapsed)
					successes++
				}
				if successes == 0 {
					log.Error("ComputeBestMaskDutyOneOrderSync: all runs failed")
				} else {
					var total time.Duration
					for _, d := range durations {
						total += d
					}
					avg := time.Duration(int64(total) / int64(successes))
					log.WithFields(log.Fields{
						"runs":         successes,
						"avg_cost":     avg.String(),
						"attack_count": attackDutiesCount,
					}).Info("ComputeBestMaskDutyOneOrderSync average timing")
				}
			}
		case 2:
			{
				// run ComputeBestMaskDutyOneOrderMultiProcess testCount times and record durations
				durations := make([]time.Duration, 0, testCount)
				successes := 0

				for i := 0; i < testCount; i++ {
					tStart := time.Now()
					_, err := ComputeBestMaskDutyOneOrderMultiProcess(seed, allRandaoReveal, cState.Reset(), uint64(curSlot), int64(epoch), attackDuties, validatorSortedIndex, fullOrder[0])
					elapsed := time.Since(tStart)
					if err != nil {
						log.WithFields(log.Fields{"err": err, "iteration": i}).Error("ComputeBestMaskDutyOneOrderMultiProcess failed")
						// continue to next iteration, don't include failed run in averages
						continue
					}
					durations = append(durations, elapsed)
					successes++
				}
				if successes == 0 {
					log.Error("ComputeBestMaskDutyOneOrderMultiProcess: all runs failed")
				} else {
					var total time.Duration
					for _, d := range durations {
						total += d
					}
					avg := time.Duration(int64(total) / int64(successes))
					log.WithFields(log.Fields{
						"runs":         successes,
						"avg_cost":     avg.String(),
						"attack_count": attackDutiesCount,
					}).Info("ComputeBestMaskDutyOneOrderMultiProcess average timing")
				}
			}
		case 3:
			{
				// run ComputeBestMaskDutyFullTime testCount times and record durations
				durations := make([]time.Duration, 0, testCount)
				successes := 0
				for i := 0; i < testCount/20; i++ {
					tStart := time.Now()
					_, err := ComputeBestMaskDutyFullTime(seed, allRandaoReveal, cState.Reset(), uint64(curSlot), int64(epoch), attackDuties, validatorSortedIndex, fullOrder)
					elapsed := time.Since(tStart)
					if err != nil {
						log.WithFields(log.Fields{"err": err, "iteration": i}).Error("ComputeBestMaskDutyFullTime failed")
						// continue to next iteration, don't include failed run in averages
						continue
					}
					durations = append(durations, elapsed)
					successes++
				}
				if successes == 0 {
					log.Error("ComputeBestMaskDutyFullTime: all runs failed")
				} else {
					var total time.Duration
					for _, d := range durations {
						total += d
					}
					avg := time.Duration(int64(total) / int64(successes))
					log.WithFields(log.Fields{
						"runs":         successes,
						"avg_cost":     avg.String(),
						"attack_count": attackDutiesCount,
					}).Info("ComputeBestMaskDutyFullTime average timing")
				}
			}
		}

	}

}

type ValidatorInfo struct {
	PK     string `json:"pk"`
	PubKey string `json:"pubkey"`
	Index  int    `json:"index"`
}

func getValidatorListFromFile(filePath string) (map[string]ValidatorInfo, error) {
	validators := make(map[string]ValidatorInfo)
	var list []ValidatorInfo
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	for _, v := range list {
		validators[v.PubKey] = v
	}
	return validators, nil
}

func toInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func GetAllRandaoReveal(mostate *disguisedRandao.MoState, epoch int64, currentDuty []types.ProposerDuty, validatorList []ValidatorInfo) (map[string][]byte, error) {
	t1 := time.Now()
	allRandao := make(map[string][]byte)
	dv, err := mostate.RandaoDomainData(primitives.Epoch(epoch))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get randao domain data")
		return nil, err
	}
	for i := 0; i < len(currentDuty); i++ {
		duty := currentDuty[i]
		keyInfo := validatorList[toInt(duty.ValidatorIndex)]
		randaoReveal, err := GenerateRandaoRevealWithDv(dv, keyInfo.PK, keyInfo.PubKey, primitives.Epoch(epoch))
		if err != nil {
			return nil, err
		}
		allRandao[keyInfo.PubKey] = randaoReveal
	}
	t2 := time.Now()
	log.WithFields(log.Fields{
		"cost": t2.Sub(t1).String(),
	}).Debug("all randao reveal")
	return allRandao, nil
}

func ComputeBestMaskDutyOneOrderSync(allRandao map[string][]byte, cState *disguisedRandao.MoState, slot uint64, epoch int64, currentDuty []types.ProposerDuty, validatorList []ValidatorInfo, order []int) (types.ProposerDuty, error) {
	currentEpoch := epoch
	next2Epoch := currentEpoch + 2
	{
		t1 := time.Now()
		// first compute a maskInfo when don't mask any slot.
		for idx, skip := range order {
			if skip == 1 {
				continue
			}
			duty := currentDuty[idx]

			keyInfo := validatorList[toInt(duty.ValidatorIndex)]
			randaoReveal := allRandao[keyInfo.PubKey]

			if err := disguisedRandao.ProcessRandaoNoVerify(cState, randaoReveal, primitives.Epoch(currentEpoch)); err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("failed to process randao reveal")
				return types.ProposerDuty{}, err
			}
		}
		t2 := time.Now()
		log.WithFields(log.Fields{
			"process all randao": t2.Sub(t1).String(),
		}).Debug("liveness attack strategy processed all randao reveals")

		// epoch process.
		_, _, err := PrecomputeProposerIndicesSync(cState, allIndices, primitives.Epoch(next2Epoch))
		if err != nil {
			log.WithFields(log.Fields{
				"current": currentEpoch,
			}).Error("failed to precompute proposer indices")
			return types.ProposerDuty{}, err
		}
		t3 := time.Now()
		log.WithFields(log.Fields{
			"PrecomputeProposerIndicesInSync": t3.Sub(t2).String(),
			"Total Cost":                      t3.Sub(t1).String(),
		}).Debug("liveness attack precompute proposer indices timing")
	}
	return types.ProposerDuty{}, nil
}

// --- New reusable proposer worker pool ---
type proposerTask struct {
	seed          [32]byte
	validatorList []*phase0.Validator
	activeIndices []primitives.ValidatorIndex
	idx           int
	slot          uint64
	resCh         chan proposerResult
}

type proposerResult struct {
	idx   int
	index primitives.ValidatorIndex
	err   error
}

type ProposerWorkerPool struct {
	tasks       chan proposerTask
	workerCount int
	stopCh      chan struct{}
}

func NewProposerWorkerPool(workerCnt int) *ProposerWorkerPool {
	p := &ProposerWorkerPool{
		tasks:       make(chan proposerTask, 1024),
		workerCount: workerCnt,
		stopCh:      make(chan struct{}),
	}
	for i := 0; i < workerCnt; i++ {
		go p.worker()
	}
	return p
}

func (p *ProposerWorkerPool) worker() {
	for {
		select {
		case t := <-p.tasks:
			// compute seedWithSlot
			seedWithSlot := append(t.seed[:], bytesutil.Bytes8(uint64(t.slot)+uint64(t.idx))...)
			// compute hash of seedWithSlot deterministically
			var seedHash [32]byte
			h := sha256.Sum256(seedWithSlot)
			copy(seedHash[:], h[:])
			index, err := ComputeProposerIndex(t.validatorList, t.activeIndices, t.seed)
			// send result back (non-blocking in case caller gave buffer)
			t.resCh <- proposerResult{idx: t.idx, index: index, err: err}
		case <-p.stopCh:
			return
		}
	}
}

func (p *ProposerWorkerPool) Submit(t proposerTask) {
	p.tasks <- t
}

var proposerPool *ProposerWorkerPool
var proposerPoolOnce sync.Once

func ensureProposerPool() {
	proposerPoolOnce.Do(func() {
		// use number of CPU cores as default worker count
		wc := runtime.NumCPU()
		if wc <= 0 {
			wc = 4
		}
		log.WithField("worker count", wc).Info("Initializing proposer worker pool")
		proposerPool = NewProposerWorkerPool(wc)
	})
}

func ComputeBestMaskDutyOneOrderMultiProcess(seed [32]byte, allRandao map[string][]byte, cState *disguisedRandao.MoState, slot uint64, epoch int64, currentDuty []types.ProposerDuty, validatorList []ValidatorInfo, order []int) (types.ProposerDuty, error) {
	currentEpoch := epoch
	next2Epoch := currentEpoch + 2
	{
		t1 := time.Now()
		// first compute a maskInfo when don't mask any slot.
		for idx, skip := range order {
			if skip == 1 {
				continue
			}
			duty := currentDuty[idx]
			keyInfo := validatorList[toInt(duty.ValidatorIndex)]
			randaoReveal := allRandao[keyInfo.PubKey]

			if err := disguisedRandao.ProcessRandaoNoVerify(cState, randaoReveal, primitives.Epoch(currentEpoch)); err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("failed to process randao reveal")
				return types.ProposerDuty{}, err
			}
		}
		t2 := time.Now()
		log.WithFields(log.Fields{
			"process all randao": t2.Sub(t1).String(),
		}).Debug("liveness attack strategy processed all randao reveals")

		// epoch process.
		_, _, err := PrecomputeProposerIndicesMultiProcess(seed, cState.ValidatorList(), allIndices, primitives.Epoch(next2Epoch))
		if err != nil {
			log.WithFields(log.Fields{
				"current": currentEpoch,
			}).Error("failed to precompute proposer indices")
			return types.ProposerDuty{}, err
		}
		t3 := time.Now()
		log.WithFields(log.Fields{
			"PrecomputeProposerIndicesMulti": t3.Sub(t2).String(),
			"Total Cost":                     t3.Sub(t1).String(),
		}).Debug("liveness attack precompute proposer indices timing")

	}
	return types.ProposerDuty{}, nil
}

func ComputeBestMaskDutyFullTime(seed [32]byte, allRandao map[string][]byte, cState *disguisedRandao.MoState, slot uint64, epoch int64, currentDuty []types.ProposerDuty, validatorList []ValidatorInfo, fullOrder [][]int) (types.ProposerDuty, error) {
	t1 := time.Now()
	for _, order := range fullOrder {
		ComputeBestMaskDutyOneOrderMultiProcess(seed, allRandao, cState.Reset(), uint64(slot), int64(epoch), currentDuty, validatorList, order)
	}

	t2 := time.Now()
	log.WithFields(log.Fields{
		"total cost":  t2.Sub(t1).String(),
		"order count": len(fullOrder),
	}).Debug("ComputeBestMaskDutyFullTime")

	return types.ProposerDuty{}, nil
}

func GenerateRandaoRevealWithDv(dv []byte, privk string, pubkey string, epoch primitives.Epoch) ([]byte, error) {
	sszUint := primitives.SSZUint64(epoch)
	root, err := signing.ComputeSigningRoot(&sszUint, dv)
	if err != nil {
		return nil, err
	}
	// private key to private key.
	secretKey, err := bls.SecretKeyFromBytes(common.FromHex(privk))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize keys privk, err:%s", err.Error())
	}
	randaoReveal := secretKey.Sign(root[:])
	return randaoReveal.Marshal(), nil
}

func PrecomputeProposerIndicesSync(state *disguisedRandao.MoState, activeIndices []primitives.ValidatorIndex, e primitives.Epoch) ([]byte, []primitives.ValidatorIndex, error) {
	hashFunc := hash.CustomSHA256Hasher()
	proposerIndices := make([]primitives.ValidatorIndex, 32)

	seed, err := disguisedRandao.Seed(state, e, disguisedRandao.DomainBeaconProposer)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not generate seed")
	}
	slot := e * 32
	for i := uint64(0); i < uint64(32); i++ {
		seedWithSlot := append(seed[:], bytesutil.Bytes8(uint64(slot)+i)...)
		seedWithSlotHash := hashFunc(seedWithSlot)
		index, err := ComputeProposerIndex(state.ValidatorList(), activeIndices, seedWithSlotHash)
		if err != nil {
			return nil, nil, err
		}
		proposerIndices[i] = index
	}

	return seed[:], proposerIndices, nil
}

func PrecomputeProposerIndicesMultiProcess(seed [32]byte, validators []*phase0.Validator, activeIndices []primitives.ValidatorIndex, e primitives.Epoch) ([]byte, []primitives.ValidatorIndex, error) {
	proposerIndices := make([]primitives.ValidatorIndex, 32)
	slot := e * 32

	// use a reusable worker pool instead of spawning new goroutines each call
	ensureProposerPool()

	resCh := make(chan proposerResult, 32)
	// submit tasks to the pool
	for i := uint64(0); i < uint64(32); i++ {
		ii := i

		task := proposerTask{
			seed:          seed,
			slot:          uint64(slot),
			validatorList: validators,
			activeIndices: activeIndices,
			idx:           int(ii),
			resCh:         resCh,
		}
		proposerPool.Submit(task)
	}

	// collect results or return on first error
	received := 0
	for received < 32 {
		r := <-resCh
		if r.err != nil {
			return nil, nil, r.err
		}
		proposerIndices[r.idx] = r.index
		received++
	}

	return seed[:], proposerIndices, nil
}

func ComputeProposerIndex(validatorList []*phase0.Validator, activeIndices []primitives.ValidatorIndex, seed [32]byte) (primitives.ValidatorIndex, error) { // luxq: go with here.
	length := uint64(len(activeIndices))
	if length == 0 {
		return 0, errors.New("empty active indices list")
	}
	maxRandomByte := uint64(1<<8 - 1)
	hashFunc := hash.CustomSHA256Hasher()

	for i := uint64(0); ; i++ {
		candidateIndex, err := helpers.ComputeShuffledIndex(primitives.ValidatorIndex(i%length), length, seed, true /* shuffle */)
		if err != nil {
			return 0, err
		}
		candidateIndex = activeIndices[candidateIndex]
		if uint64(candidateIndex) >= uint64(len(validatorList)) { // total number.
			return 0, errors.New("active index out of range")
		}
		b := append(seed[:], bytesutil.Bytes8(i/32)...)
		randomByte := hashFunc(b)[i%32]
		v := validatorList[candidateIndex]
		if v == nil {
			return 0, errors.New("nil validator")
		}
		effectiveBal := uint64(v.EffectiveBalance) // 32 * 10 ** 18

		maxEB := disguisedRandao.MaxEffectiveBalance

		if effectiveBal*maxRandomByte >= maxEB*uint64(randomByte) {
			return candidateIndex, nil
		} else {
		}
	}
}

// RandomSamplePreserveOrder selects k elements from arr preserving their original order.
// Selection is uniform over all combinations of k elements from n (if k <= n).
// If k <= 0 returns an empty slice. If k >= len(arr) returns a copy of arr.
// A rand.Source may be provided via r; if r is nil, a new rand.Rand with current time seed is used.
func RandomAttackerDuties(r *rand.Rand, arr []types.ProposerDuty, k int) []types.ProposerDuty {
	n := len(arr)
	if k <= 0 || n == 0 {
		return nil
	}
	if k >= n {
		out := make([]types.ProposerDuty, n)
		copy(out, arr)
		return out
	}
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	result := make([]types.ProposerDuty, 0, k)
	need := k
	for i := 0; i < n; i++ {
		remaining := n - i
		// choose current element with probability need/remaining
		// use Intn(remaining) < need to achieve that
		if r.Intn(remaining) < need {
			result = append(result, arr[i])
			need--
			if need == 0 {
				break
			}
		}
	}
	return result
}

func GenerateRandomDuty(validators []ValidatorInfo, epoch int) []types.ProposerDuty {
	startSlot := epoch * 32
	randomSeq := GetRandomUniqueNumbers()
	duties := make([]types.ProposerDuty, 0)
	for i, idx := range randomSeq {
		duty := types.ProposerDuty{
			ValidatorIndex: strconv.Itoa(idx),
			Slot:           strconv.Itoa(startSlot + i),
			Pubkey:         validators[idx].PubKey,
		}
		duties = append(duties, duty)
	}
	return duties
}

func GetGenesisTime(state *spec.VersionedBeaconState) int64 {
	if state.Phase0 != nil {
		return int64(state.Phase0.GenesisTime)
	}
	if state.Altair != nil {
		return int64(state.Altair.GenesisTime)
	}
	if state.Bellatrix != nil {
		return int64(state.Bellatrix.GenesisTime)
	}
	if state.Capella != nil {
		return int64(state.Capella.GenesisTime)
	}
	if state.Deneb != nil {
		return int64(state.Deneb.GenesisTime)
	}
	return 0
}
