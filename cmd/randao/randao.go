package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"github.com/prysmaticlabs/prysm/v5/crypto/hash"
	"github.com/prysmaticlabs/prysm/v5/encoding/bytesutil"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/disguisedRandao"
	"github.com/tsinghua-cel/attacker-service/types"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var (
	beaconUrl            = flag.String("beacon-url", "", "Beacon URL")
	stateFile            = flag.String("state", "", "State file")
	validatorList        = flag.String("validator-list", "", "Validator list file path")
	slot                 = flag.String("slot", "161", "Slot to test")
	testCase             = flag.Int("case", 1, "Test case number(1,2,3")
	fullTimeRoutineCount = flag.Int("routine", 32, "Routine count that calculate full time used")
)

var (
	allIndices = disguisedRandao.GenValidatorIndices(0, 255)
)

func main() {
	flag.Parse()
	beaconClient := beaconapi.NewBeaconGwClient(*beaconUrl)
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
		chainstate, err := beaconClient.GetBeaconState(*slot)
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
	curSlot, _ := state.Slot()
	epoch := curSlot / 32

	proposerDuty, err := beaconClient.GetEpochProposerDuties(int(epoch))
	if err != nil {
		log.WithFields(log.Fields{
			"epoch": epoch,
			"err":   err,
		}).Fatal("failed to get proposer duties")
		return
	}
	type TestSetting struct {
		attackCount int
		testCount   int
	}
	var testSet = []TestSetting{
		{10, 50},
		{20, 50},
	}
	for _, setting := range testSet {
		attackDutiesCount := setting.attackCount
		testCount := setting.testCount
		attackDuties := RandomAttackerDuties(rand.New(rand.NewSource(time.Now().UnixNano())), proposerDuty, attackDutiesCount)
		t1 := time.Now()
		allRandaoReveal, _ := GetAllRandaoReveal(mostate, int64(epoch), attackDuties, validatorSortedIndex)
		t2 := time.Now()
		log.WithFields(log.Fields{
			"cost":         t2.Sub(t1).String(),
			"attack_count": attackDutiesCount,
		}).Info("get all randao reveal for attacker duties")
		fullOrder := GetAllBinarySequences(attackDutiesCount)

		log.Println("full order count ", len(fullOrder))
		switch *testCase {
		case 1:
			{
				// run ComputeBestMaskDutyOneOrderSync testCount times and record durations
				durations := make([]time.Duration, 0, testCount)
				successes := 0
				for i := 0; i < testCount; i++ {
					cState, _ := disguisedRandao.InitMoState(*state)
					tStart := time.Now()
					_, err := ComputeBestMaskDutyOneOrderSync(allRandaoReveal, cState, uint64(curSlot), int64(epoch), attackDuties, validatorSortedIndex, fullOrder[0])
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
					cState, _ := disguisedRandao.InitMoState(*state)
					tStart := time.Now()
					_, err := ComputeBestMaskDutyOneOrderMultiProcess(allRandaoReveal, cState, uint64(curSlot), int64(epoch), attackDuties, validatorSortedIndex, fullOrder[0])
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
				for i := 0; i < testCount/10; i++ {
					cState, _ := disguisedRandao.InitMoState(*state)
					tStart := time.Now()
					_, err := ComputeBestMaskDutyFullTime(allRandaoReveal, cState, uint64(curSlot), int64(epoch), attackDuties, validatorSortedIndex, fullOrder)
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

func ComputeBestMaskDutyOneOrderMultiProcess(allRandao map[string][]byte, cState *disguisedRandao.MoState, slot uint64, epoch int64, currentDuty []types.ProposerDuty, validatorList []ValidatorInfo, order []int) (types.ProposerDuty, error) {
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
		_, _, err := PrecomputeProposerIndicesMultiProcess(cState, allIndices, primitives.Epoch(next2Epoch))
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

func ComputeBestMaskDutyFullTime(allRandao map[string][]byte, cState *disguisedRandao.MoState, slot uint64, epoch int64, currentDuty []types.ProposerDuty, validatorList []ValidatorInfo, fullOrder [][]int) (types.ProposerDuty, error) {
	t1 := time.Now()
	workerCount := *fullTimeRoutineCount
	// dispatch orders to 32 worker goroutines
	jobs := make(chan []int)
	errCh := make(chan error, 1)
	done := make(chan struct{}, workerCount)
	workers := workerCount

	for i := 0; i < workers; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for order := range jobs {
				if _, err := ComputeBestMaskDutyOneOrderSync(allRandao, cState.Clone(), uint64(slot), int64(epoch), currentDuty, validatorList, order); err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
			}
		}()
	}

	// send jobs, stop early if an error is reported
	var finalErr error
SendLoop:
	for _, order := range fullOrder {
		select {
		case e := <-errCh:
			finalErr = e
			break SendLoop
		case jobs <- order:
		}
	}
	close(jobs)

	// wait for all workers to finish
	for i := 0; i < workers; i++ {
		<-done
	}

	// check for any error reported after workers finished
	if finalErr == nil {
		select {
		case e := <-errCh:
			finalErr = e
		default:
		}
	}
	if finalErr != nil {
		log.WithFields(log.Fields{
			"err": finalErr,
		}).Error("ComputeBestMaskDutyOneOrderSync failed")
		return types.ProposerDuty{}, finalErr
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

func PrecomputeProposerIndicesMultiProcess(state *disguisedRandao.MoState, activeIndices []primitives.ValidatorIndex, e primitives.Epoch) ([]byte, []primitives.ValidatorIndex, error) {
	hashFunc := hash.CustomSHA256Hasher()
	proposerIndices := make([]primitives.ValidatorIndex, 32)

	seed, err := disguisedRandao.Seed(state, e, disguisedRandao.DomainBeaconProposer)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not generate seed")
	}
	slot := e * 32

	// run proposer index computation concurrently
	errCh := make(chan error, 1)
	resCh := make(chan struct {
		idx   int
		index primitives.ValidatorIndex
	}, 32)

	for i := uint64(0); i < uint64(32); i++ {
		ii := i
		go func() {
			seedWithSlot := append(seed[:], bytesutil.Bytes8(uint64(slot)+ii)...)
			seedWithSlotHash := hashFunc(seedWithSlot)
			index, err := ComputeProposerIndex(state.ValidatorList(), activeIndices, seedWithSlotHash)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
			resCh <- struct {
				idx   int
				index primitives.ValidatorIndex
			}{int(ii), index}
		}()
	}

	// collect results or return on first error
	received := 0
	for received < 32 {
		select {
		case err := <-errCh:
			return nil, nil, err
		case r := <-resCh:
			proposerIndices[r.idx] = r.index
			received++
		}
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
