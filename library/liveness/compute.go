package liveness

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"github.com/prysmaticlabs/prysm/v5/crypto/hash"
	"github.com/prysmaticlabs/prysm/v5/encoding/bytesutil"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/disguisedRandao"
	"github.com/tsinghua-cel/attacker-service/types"
	"runtime"
	"sync"
	"time"
)

var (
	allIndices = disguisedRandao.GenValidatorIndices(0, 255)
)

func (o *Instance) ComputeBestMask(slot uint64, currentFullDuty []types.ProposerDuty, cmper ProposerListCmper) (BestMaskDutyInfo, error) {
	currentState, err := o.b.GetBeaconState("head")
	if err != nil {
		log.WithFields(log.Fields{
			"paramSlot": slot,
			"err":       err,
		}).Error("failed to get beacon state")
		return BestMaskDutyInfo{}, err
	}
	var state = currentState

	chainValidators, err := state.Validators()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("failed to get validators from beacon state")
		return BestMaskDutyInfo{}, err
	}
	validatorInfos := o.b.GetAllValidatorsInfo()

	validatorSortedIndex := make([]types.ValidatorKeysInfo, len(chainValidators))
	for index, v := range chainValidators {
		pubkey := v.PublicKey.String()
		if info, ok := validatorInfos[pubkey]; ok {
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
		return BestMaskDutyInfo{}, err
	}

	curSlot, _ := state.Slot()
	epoch := common.SlotToEpoch(int64(curSlot))
	attackDuties := o.param.FilterHackerDuties(currentFullDuty)

	attackDutiesCount := len(attackDuties)
	t1 := time.Now()
	// attacker duty : slot-32, slot-35, slot-48 .... slot-63
	// order:             0         1       0     ..     0

	// 1 is mask, 0 is not mask, only consider the case that mask 5 slots.
	fullOrder := common.GetAllBinarySequencesWithMaxOnes(attackDutiesCount, 5)
	targetEpoch := primitives.Epoch(epoch + 2)
	cState := mostate.Reset()

	seed, _ := disguisedRandao.Seed(cState, targetEpoch, disguisedRandao.DomainBeaconProposer)
	allRandaoReveal, _ := GetAllRandaoReveal(cState, int64(epoch), attackDuties, validatorSortedIndex)
	t2 := time.Now()
	log.WithFields(log.Fields{
		"cost":         t2.Sub(t1).String(),
		"attack_count": attackDutiesCount,
		"order_count":  len(fullOrder),
	}).Debug("get all randao reveal for attacker maskedDuties")

	log.WithFields(log.Fields{
		"cost":         t2.Sub(t1).String(),
		"attack_count": attackDutiesCount,
		"order_count":  len(fullOrder),
		"seed":         hex.EncodeToString(seed[:]),
		"curSlot":      curSlot,
		"epoch":        epoch,
	}).Debug("goto compute best mask duty for attacker maskedDuties")

	return o.ComputeBestMaskDutyFullTime(seed, allRandaoReveal, cState.Reset(), uint64(curSlot), int64(epoch), attackDuties, validatorSortedIndex, fullOrder, cmper)
}

func GetAllRandaoReveal(mostate *disguisedRandao.MoState, epoch int64, currentDuty []types.ProposerDuty, validatorList []types.ValidatorKeysInfo) (map[string][]byte, error) {
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
		randaoReveal, err := GenerateRandaoRevealWithDv(dv, keyInfo.Private, keyInfo.Pubkey, primitives.Epoch(epoch))
		if err != nil {
			return nil, err
		}
		allRandao[keyInfo.Pubkey] = randaoReveal
	}
	t2 := time.Now()
	log.WithFields(log.Fields{
		"cost": t2.Sub(t1).String(),
	}).Debug("all randao reveal")
	return allRandao, nil
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

func ComputeBestMaskDutyOneOrderMultiProcess(seed [32]byte, allRandao map[string][]byte,
	cState *disguisedRandao.MoState, slot uint64, epoch int64, currentDuty []types.ProposerDuty,
	validatorList []types.ValidatorKeysInfo, order []int) ([]primitives.ValidatorIndex, error) {

	currentEpoch := epoch
	next2Epoch := currentEpoch + 2

	t1 := time.Now()
	// order [0:31]
	for idx, skip := range order {
		duty := currentDuty[idx]
		if skip == 1 || toInt(duty.Slot) <= int(slot) {
			continue
		}

		keyInfo := validatorList[toInt(duty.ValidatorIndex)]
		randaoReveal := allRandao[keyInfo.Pubkey]

		if err := disguisedRandao.ProcessRandaoNoVerify(cState, randaoReveal, primitives.Epoch(currentEpoch)); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to process randao reveal")
			return []primitives.ValidatorIndex{}, err
		}
	}
	t2 := time.Now()
	log.WithFields(log.Fields{
		"process all randao": t2.Sub(t1).String(),
	}).Debug("liveness attack strategy processed all randao reveals")

	// epoch process.
	_, proposers, err := PrecomputeProposerIndicesMultiProcess(seed, cState.ValidatorList(), allIndices, primitives.Epoch(next2Epoch))
	if err != nil {
		log.WithFields(log.Fields{
			"current": currentEpoch,
		}).Error("failed to precompute proposer indices")
		return []primitives.ValidatorIndex{}, err
	}
	t3 := time.Now()
	log.WithFields(log.Fields{
		"PrecomputeProposerIndicesMulti": t3.Sub(t2).String(),
		"Total Cost":                     t3.Sub(t1).String(),
	}).Debug("liveness attack precompute proposer indices timing")

	return proposers, nil
}

func (o *Instance) ComputeBestMaskDutyFullTime(seed [32]byte, allRandao map[string][]byte,
	cState *disguisedRandao.MoState, slot uint64, epoch int64, attackDuties []types.ProposerDuty,
	validatorList []types.ValidatorKeysInfo, fullOrder [][]int, cmper ProposerListCmper) (BestMaskDutyInfo, error) {
	t1 := time.Now()
	var bestMask BestMaskDutyInfo
	for _, order := range fullOrder {
		resultProserList, err := ComputeBestMaskDutyOneOrderMultiProcess(seed, allRandao, cState.Reset(), uint64(slot), int64(epoch), attackDuties, validatorList, order)
		if err != nil {
			log.WithFields(log.Fields{
				"err":   err,
				"order": order,
			}).Error("failed to compute best mask duty for one order")
			continue
		}
		curMaskInfo := BestMaskDutyInfo{
			FirstIsAttack:  o.param.IsHackValidator(int(resultProserList[0])),
			AttackersCount: o.attackerCount(resultProserList),
			order:          order,
			proposers:      resultProserList,
			seed:           seed[:],
		}
		if cmper.BetterThan(curMaskInfo, bestMask) {
			bestMask = curMaskInfo
		}
	}

	t2 := time.Now()
	log.WithFields(log.Fields{
		"total cost":  t2.Sub(t1).String(),
		"order count": len(fullOrder),
	}).Debug("ComputeBestMaskDutyFullTime")
	// fill masked maskedDuties for best mask.
	var maskedDuties []types.ProposerDuty
	for idx, skip := range bestMask.order {
		if skip == 1 {
			maskedDuties = append(maskedDuties, attackDuties[idx])
		}
	}
	bestMask.maskedDuties = maskedDuties
	bestMask.attackerDuties = attackDuties

	return bestMask, nil
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
		return nil, nil, fmt.Errorf("could not generate seed, err:%s", err)
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
