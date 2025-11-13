package disguisedRandao

import (
	"encoding/hex"
	"fmt"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/state/state-native/types"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	multi_value_slice "github.com/prysmaticlabs/prysm/v5/container/multi-value-slice"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"github.com/prysmaticlabs/prysm/v5/crypto/hash"
	"github.com/prysmaticlabs/prysm/v5/encoding/bytesutil"
	eth "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v5/runtime/version"
	"github.com/prysmaticlabs/prysm/v5/time/slots"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"sync"
)

var (
	MaxEffectiveBalance       uint64           = 32 * (10 ^ 9)                   // Gwei
	DomainBeaconProposer      [4]byte          = [4]byte{0x00, 0x00, 0x00, 0x00} // Domain for beacon proposer
	DomainRandao              [4]byte          = [4]byte{0x02, 0x00, 0x00, 0x00}
	EpochsPerHistoricalVector primitives.Epoch = 65536
	MinSeedLookahead          primitives.Epoch = 1 // todo:check the value.
)

type MoState struct {
	id                    uint64
	lock                  sync.RWMutex
	version               int
	genesisTime           uint64
	genesisValidatorsRoot [32]byte
	slot                  primitives.Slot
	fork                  *phase0.Fork
	validators            []*phase0.Validator
	historyRanDaoMix      map[primitives.Epoch]phase0.Root
	bakHistoryRanDaoMix   map[primitives.Epoch]phase0.Root
	curAheadEpoch         primitives.Epoch
	curEpoch              primitives.Epoch
	targetAheadEpoch      primitives.Epoch
	targetEpoch           primitives.Epoch
}

func (b *MoState) Id() multi_value_slice.Id {
	return b.id
}

func InitMoState(beaconState spec.VersionedBeaconState) (*MoState, error) {
	state := beaconState.Deneb
	moState := &MoState{
		id:                    types.Enumerator.Inc(),
		version:               version.Deneb,
		genesisTime:           state.GenesisTime,
		genesisValidatorsRoot: state.GenesisValidatorsRoot,
		slot:                  primitives.Slot(state.Slot),
		fork:                  state.Fork,
		historyRanDaoMix:      make(map[primitives.Epoch]phase0.Root),
		bakHistoryRanDaoMix:   make(map[primitives.Epoch]phase0.Root),
		curEpoch:              primitives.Epoch(common.SlotToEpoch(int64(state.Slot))),
		targetEpoch:           primitives.Epoch(common.SlotToEpoch(int64(state.Slot))) + 2,
	}
	//randaoMixValues := make([][]byte, len(state.RANDAOMixes))
	//for i, mix := range state.RANDAOMixes {
	//	randaoMixValues[i] = make([]byte, 32)
	//	copy(randaoMixValues[i], mix[:])
	//}
	//historyRandaoMixValues := state_native.NewMultiValueRandaoMixes(randaoMixValues)
	curAheadEpoch := moState.curEpoch + EpochsPerHistoricalVector -
		MinSeedLookahead - 1

	moState.curAheadEpoch = curAheadEpoch

	targetAheadEpoch := moState.targetEpoch + EpochsPerHistoricalVector -
		MinSeedLookahead - 1
	moState.targetAheadEpoch = targetAheadEpoch

	moState.historyRanDaoMix[moState.curAheadEpoch] = state.RANDAOMixes[targetAheadEpoch%EpochsPerHistoricalVector]
	moState.bakHistoryRanDaoMix[moState.curAheadEpoch] = state.RANDAOMixes[targetAheadEpoch%EpochsPerHistoricalVector]

	moState.historyRanDaoMix[moState.targetAheadEpoch] = state.RANDAOMixes[targetAheadEpoch%EpochsPerHistoricalVector]
	moState.bakHistoryRanDaoMix[moState.targetAheadEpoch] = state.RANDAOMixes[targetAheadEpoch%EpochsPerHistoricalVector]

	moState.historyRanDaoMix[moState.curEpoch] = state.RANDAOMixes[moState.curEpoch%EpochsPerHistoricalVector]
	moState.bakHistoryRanDaoMix[moState.curEpoch] = state.RANDAOMixes[moState.curEpoch%EpochsPerHistoricalVector]

	moState.historyRanDaoMix[moState.targetEpoch] = state.RANDAOMixes[moState.targetEpoch%EpochsPerHistoricalVector]
	moState.bakHistoryRanDaoMix[moState.targetEpoch] = state.RANDAOMixes[moState.targetEpoch%EpochsPerHistoricalVector]

	moState.validators = make([]*phase0.Validator, len(state.Validators))
	for i, v := range state.Validators {
		if v == nil {
			moState.validators[i] = nil
		} else {
			moState.validators[i] = v
		}
	}

	return moState, nil
}

func (b *MoState) Dump() {
	log.WithFields(log.Fields{
		"slot":        b.slot,
		"curEpoch":    b.curEpoch,
		"targetEpoch": b.targetEpoch,
		"curAhead":    b.curAheadEpoch,
		"targetAhead": b.targetAheadEpoch,
		"randaoMixes": b.historyRanDaoMix,
	}).Info("MoState Dump")
}

func (b *MoState) Reset() *MoState {
	for k, v := range b.bakHistoryRanDaoMix {
		b.historyRanDaoMix[k] = v
	}
	return b
}

// UpdateRandaoMixesAtIndex for the beacon state. Updates the randao mixes
// at a specific index to a new value.
func (b *MoState) UpdateRandaoMixesAtIndex(epoch primitives.Epoch, val [32]byte) error {
	if _, exist := b.historyRanDaoMix[epoch]; !exist {
		return fmt.Errorf("randao mixes for epoch %d do not exist, current mixes: %+v", epoch, b.historyRanDaoMix)
	}
	b.historyRanDaoMix[epoch] = val
	return nil
}

// RandaoMixAtIndex retrieves a specific block root based on an
// input index value.
func (b *MoState) RandaoMixAtIndex(epoch primitives.Epoch) ([]byte, error) {
	if v, exist := b.historyRanDaoMix[epoch]; !exist {
		return nil, fmt.Errorf("randao mixes for epoch %d do not exist, current mixes: %+v", epoch, b.historyRanDaoMix)
	} else {
		return v[:], nil
	}
}

// PrecomputeProposerIndices computes proposer indices of the current epoch and returns a list of proposer indices,
// the index of the list represents the slot number.
func (b *MoState) PrecomputeProposerIndices(activeIndices []primitives.ValidatorIndex, e primitives.Epoch) ([]byte, []primitives.ValidatorIndex, error) {
	hashFunc := hash.CustomSHA256Hasher()
	proposerIndices := make([]primitives.ValidatorIndex, common.GetChainBaseInfo().SlotsPerEpoch)

	seed, err := Seed(b, e, DomainBeaconProposer)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not generate seed")
	}
	slot, err := slots.EpochStart(e)
	if err != nil {
		return nil, nil, err
	}
	for i := uint64(0); i < uint64(common.GetChainBaseInfo().SlotsPerEpoch); i++ {
		seedWithSlot := append(seed[:], bytesutil.Bytes8(uint64(slot)+i)...)
		seedWithSlotHash := hashFunc(seedWithSlot)
		index, err := ComputeProposerIndex(b, activeIndices, seedWithSlotHash)
		if err != nil {
			return nil, nil, err
		}
		log.WithFields(log.Fields{
			"epoch":        e,
			"slot":         uint64(slot) + i,
			"stateSlot":    b.slot,
			"valIndex":     index,
			"seed":         hex.EncodeToString(seed[:]),
			"seedWithSlot": hex.EncodeToString(seedWithSlotHash[:]),
		}).Debug("PrecomputeProposerIndices - compute proposer")
		proposerIndices[i] = index
	}

	return seed[:], proposerIndices, nil
}

func (b *MoState) RandaoDomainData(epoch primitives.Epoch) ([]byte, error) {
	var ethFork = eth.Fork{
		PreviousVersion: b.fork.PreviousVersion[:],
		CurrentVersion:  b.fork.CurrentVersion[:],
		Epoch:           primitives.Epoch(b.fork.Epoch),
	}
	dv, err := signing.Domain(&ethFork, epoch, DomainRandao, b.genesisValidatorsRoot[:])
	log.WithFields(log.Fields{
		"previousVersion":      b.fork.PreviousVersion,
		"currentVersion":       b.fork.CurrentVersion,
		"epoch":                b.fork.Epoch,
		"genesisValidatorRoot": hexutil.Encode(b.genesisValidatorsRoot[:]),
	}).Debug("validator dump randao domain data")
	return dv, err
}

func (b *MoState) GenerateRandaoReveal(privk string, pubkey string, epoch primitives.Epoch) ([]byte, error) {
	dv, err := b.RandaoDomainData(epoch)
	if err != nil {
		return nil, err
	}
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
	log.WithFields(log.Fields{
		"epoch":        epoch,
		"domainData":   hexutil.Encode(dv),
		"pubkey":       pubkey,
		"root":         hexutil.Encode(root[:]),
		"randaoReveal": hexutil.Encode(randaoReveal.Marshal()),
	}).Debug("validator dump domain")
	return randaoReveal.Marshal(), nil
}

func (b *MoState) ValidatorList() []*phase0.Validator {
	return b.validators
}

func Seed(b *MoState, epoch primitives.Epoch, domain [bls.DomainByteLength]byte) ([32]byte, error) {
	// See https://github.com/ethereum/consensus-specs/pull/1296 for
	// rationale on why offset has to look down by 1.
	lookAheadEpoch := epoch + EpochsPerHistoricalVector -
		MinSeedLookahead - 1

	randaoMix, err := RandaoMix(b, lookAheadEpoch)
	if err != nil {
		return [32]byte{}, err
	}
	seed := append(domain[:], bytesutil.Bytes8(uint64(epoch))...)
	seed = append(seed, randaoMix...)

	seed32 := hash.Hash(seed)
	log.WithFields(log.Fields{
		"statSlot":       b.slot,
		"epoch":          epoch,
		"domain":         hex.EncodeToString(domain[:]),
		"lookAheadEpoch": lookAheadEpoch,
		"randaoMix":      hex.EncodeToString(randaoMix),
		"seed":           hex.EncodeToString(seed32[:]),
	}).Debug("Seed RandaoMix")
	return seed32, nil
}

func RandaoMix(b *MoState, epoch primitives.Epoch) ([]byte, error) {
	return b.RandaoMixAtIndex(epoch)
}

func GenValidatorIndices(from, to int) []primitives.ValidatorIndex {
	if from > to {
		return nil
	}
	indices := make([]primitives.ValidatorIndex, 0, to-from+1)
	for i := from; i <= to; i++ {
		indices = append(indices, primitives.ValidatorIndex(i))
	}
	return indices
}

func ComputeProposerIndex(bstate *MoState, activeIndices []primitives.ValidatorIndex, seed [32]byte) (primitives.ValidatorIndex, error) { // luxq: go with here.
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
		if uint64(candidateIndex) >= uint64(len(bstate.validators)) { // total number.
			return 0, errors.New("active index out of range")
		}
		b := append(seed[:], bytesutil.Bytes8(i/32)...)
		randomByte := hashFunc(b)[i%32]
		v := bstate.validators[candidateIndex]
		if v == nil {
			return 0, errors.New("nil validator")
		}
		effectiveBal := uint64(v.EffectiveBalance) // 32 * 10 ** 18

		maxEB := MaxEffectiveBalance

		if effectiveBal*maxRandomByte >= maxEB*uint64(randomByte) {
			log.WithFields(log.Fields{
				"stateSlot": bstate.slot,
				//"activeCount":  len(activeIndices),
				"proposer":           candidateIndex,
				"randomByte":         randomByte,
				"seedWithSlot":       hex.EncodeToString(seed[:]),
				"maxEB":              maxEB,
				"v.EffectiveBalance": v.EffectiveBalance,
				//"activeIndices": indicesToStr(activeIndices),
			}).Debug("compute proposer index - selected")
			return candidateIndex, nil
		} else {
			log.WithFields(log.Fields{
				"stateSlot": bstate.slot,
				//"activeCount":  len(activeIndices),
				"proposer":           candidateIndex,
				"randomByte":         randomByte,
				"seedWithSlot":       hex.EncodeToString(seed[:]),
				"maxEB":              maxEB,
				"v.EffectiveBalance": v.EffectiveBalance,
				//"activeIndices": indicesToStr(activeIndices),
			}).Debug("compute proposer index - not selected")
		}
	}
}

func ProcessRandaoNoVerify(
	beaconState *MoState,
	randaoReveal []byte,
	currentEpoch primitives.Epoch,
) error {
	// If block randao passed verification, we XOR the state's latest randao mix with the block's
	// randao and update the state's corresponding latest randao mix value.
	latestMixSlice, err := beaconState.RandaoMixAtIndex(currentEpoch)
	if err != nil {
		return err
	}
	blockRandaoReveal := hash.Hash(randaoReveal)
	if len(blockRandaoReveal) != len(latestMixSlice) {
		return errors.New("blockRandaoReveal length doesn't match latestMixSlice length")
	}
	for i, x := range blockRandaoReveal {
		latestMixSlice[i] ^= x
	}
	if err := beaconState.UpdateRandaoMixesAtIndex(currentEpoch, [32]byte(latestMixSlice)); err != nil {
		return err
	}
	//log.WithFields(log.Fields{
	//	"epoch":             currentEpoch,
	//	"randao":            hexutil.Encode(randaoReveal),
	//	"latestMixesLength": latestMixesLength,
	//	"latestMixSlice":    hexutil.Encode(latestMixSlice),
	//}).Debug("ProcessRandaoNoVerify")
	return nil
}
