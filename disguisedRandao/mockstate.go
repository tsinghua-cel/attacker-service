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
	state_native "github.com/prysmaticlabs/prysm/v5/beacon-chain/state/state-native"
	customtypes "github.com/prysmaticlabs/prysm/v5/beacon-chain/state/state-native/custom-types"
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
	randaoMixesMultiValue *state_native.MultiValueRandaoMixes
}

func (b *MoState) Id() multi_value_slice.Id {
	return b.id
}

func InitMoState(beaconState *spec.VersionedBeaconState) (*MoState, error) {
	if beaconState == nil {
		return nil, errors.New("state is nil")
	}
	state := beaconState.Deneb
	moState := &MoState{
		id:                    types.Enumerator.Inc(),
		version:               version.Deneb,
		genesisTime:           state.GenesisTime,
		genesisValidatorsRoot: state.GenesisValidatorsRoot,
		slot:                  primitives.Slot(state.Slot),
		fork:                  state.Fork,
	}
	randaoMixes := make([][]byte, len(state.RANDAOMixes))
	for i, mix := range state.RANDAOMixes {
		randaoMixes[i] = make([]byte, 32)
		copy(randaoMixes[i], mix[:])
	}
	moState.randaoMixesMultiValue = state_native.NewMultiValueRandaoMixes(randaoMixes)
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

// UpdateRandaoMixesAtIndex for the beacon state. Updates the randao mixes
// at a specific index to a new value.
func (b *MoState) UpdateRandaoMixesAtIndex(idx uint64, val [32]byte) error {
	if err := b.randaoMixesMultiValue.UpdateAt(b, idx, val); err != nil {
		return errors.Wrap(err, "could not update randao mixes")
	}

	return nil
}

// RandaoMixes of block proposers on the beacon chain.
func (b *MoState) RandaoMixes() [][]byte {
	b.lock.RLock()
	defer b.lock.RUnlock()

	mixes := b.randaoMixesVal()
	if mixes == nil {
		return nil
	}
	return mixes.Slice()
}

func (b *MoState) randaoMixesVal() customtypes.RandaoMixes {
	{
		if b.randaoMixesMultiValue == nil {
			return nil
		}
		return b.randaoMixesMultiValue.Value(b)
	}
}

// RandaoMixAtIndex retrieves a specific block root based on an
// input index value.
func (b *MoState) RandaoMixAtIndex(idx uint64) ([]byte, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	{
		if b.randaoMixesMultiValue == nil {
			return nil, nil
		}
		r, err := b.randaoMixesMultiValue.At(b, idx)
		if err != nil {
			return nil, err
		}
		return r[:], nil
	}
}

// RandaoMixesLength returns the length of the randao mixes slice.
func (b *MoState) RandaoMixesLength() int {
	b.lock.RLock()
	defer b.lock.RUnlock()

	{
		if b.randaoMixesMultiValue == nil {
			return 0
		}
		return b.randaoMixesMultiValue.Len(b)
	}
}

func (b *MoState) Clone() *MoState {
	b.lock.RLock()
	defer b.lock.RUnlock()

	clone := &MoState{
		id:                    b.id,
		version:               b.version,
		genesisTime:           b.genesisTime,
		genesisValidatorsRoot: b.genesisValidatorsRoot,
		slot:                  b.slot,
		fork:                  b.fork,
		validators:            make([]*phase0.Validator, len(b.validators)),
		randaoMixesMultiValue: state_native.NewMultiValueRandaoMixes(b.RandaoMixes()),
	}
	copy(clone.validators, b.validators)
	return clone
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
	return b.RandaoMixAtIndex(uint64(epoch % EpochsPerHistoricalVector))
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
	latestMixesLength := EpochsPerHistoricalVector
	latestMixSlice, err := beaconState.RandaoMixAtIndex(uint64(currentEpoch % latestMixesLength))
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
	if err := beaconState.UpdateRandaoMixesAtIndex(uint64(currentEpoch%latestMixesLength), [32]byte(latestMixSlice)); err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"epoch":             currentEpoch,
		"randao":            hexutil.Encode(randaoReveal),
		"latestMixesLength": latestMixesLength,
		"latestMixSlice":    hexutil.Encode(latestMixSlice),
	}).Debug("ProcessRandaoNoVerify")
	return nil
}
