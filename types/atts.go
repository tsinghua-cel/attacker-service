package types

import (
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/prysmaticlabs/prysm/v5/config/features"
	"github.com/prysmaticlabs/prysm/v5/config/params"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1/attestation"
	"github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1/attestation/aggregation"
	log "github.com/sirupsen/logrus"
	"sort"
)

type ProposerAtts []ethpb.Att

// SortByProfitability orders attestations by highest slot and by highest aggregation bit count.
func (a ProposerAtts) SortByProfitability() (ProposerAtts, error) {
	if len(a) < 2 {
		return a, nil
	}
	return a.sortByProfitabilityUsingMaxCover()
}

// sortByProfitabilityUsingMaxCover orders attestations by highest slot and by highest aggregation bit count.
// Duplicate bits are counted only once, using max-cover algorithm.
func (a ProposerAtts) sortByProfitabilityUsingMaxCover() (ProposerAtts, error) {
	// Separate attestations by slot, as slot number takes higher precedence when sorting.
	var slots []primitives.Slot
	attsBySlot := map[primitives.Slot]ProposerAtts{}
	for _, att := range a {
		if _, ok := attsBySlot[att.GetData().Slot]; !ok {
			slots = append(slots, att.GetData().Slot)
		}
		attsBySlot[att.GetData().Slot] = append(attsBySlot[att.GetData().Slot], att)
	}

	selectAtts := func(atts ProposerAtts) (ProposerAtts, error) {
		if len(atts) < 2 {
			return atts, nil
		}
		candidates := make([]*bitfield.Bitlist64, len(atts))
		for i := 0; i < len(atts); i++ {
			var err error
			candidates[i], err = atts[i].GetAggregationBits().ToBitlist64()
			if err != nil {
				return nil, err
			}
		}
		// Add selected candidates on top, those that are not selected - append at bottom.
		selectedKeys, _, err := aggregation.MaxCover(candidates, len(candidates), true /* allowOverlaps */)
		if err == nil {
			// Pick selected attestations first, leftover attestations will be appended at the end.
			// Both lists will be sorted by number of bits set.
			selectedAtts := make(ProposerAtts, selectedKeys.Count())
			leftoverAtts := make(ProposerAtts, selectedKeys.Not().Count())
			for i, key := range selectedKeys.BitIndices() {
				selectedAtts[i] = atts[key]
			}
			for i, key := range selectedKeys.Not().BitIndices() {
				leftoverAtts[i] = atts[key]
			}
			sort.Slice(selectedAtts, func(i, j int) bool {
				return selectedAtts[i].GetAggregationBits().Count() > selectedAtts[j].GetAggregationBits().Count()
			})
			sort.Slice(leftoverAtts, func(i, j int) bool {
				return leftoverAtts[i].GetAggregationBits().Count() > leftoverAtts[j].GetAggregationBits().Count()
			})
			return append(selectedAtts, leftoverAtts...), nil
		}
		return atts, nil
	}

	// Select attestations. Slots are sorted from higher to lower at this point. Within slots attestations
	// are sorted to maximize profitability (greedily selected, with previous attestations' bits
	// evaluated before including any new attestation).
	var sortedAtts ProposerAtts
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] > slots[j]
	})
	for _, slot := range slots {
		selected, err := selectAtts(attsBySlot[slot])
		if err != nil {
			return nil, err
		}
		sortedAtts = append(sortedAtts, selected...)
	}

	return sortedAtts, nil
}

// LimitToMaxAttestations limits attestations to maximum attestations per block.
func (a ProposerAtts) LimitToMaxAttestations() ProposerAtts {
	if uint64(len(a)) > params.BeaconConfig().MaxAttestations {
		return a[:params.BeaconConfig().MaxAttestations]
	}
	return a
}

func (a ProposerAtts) Sort() (ProposerAtts, error) {
	if len(a) < 2 {
		return a, nil
	}

	if features.Get().DisableCommitteeAwarePacking {
		return a.sortByProfitabilityUsingMaxCover()
	}
	return a.sortBySlotAndCommittee()
}

// sortSlotAttestations assumes each proposerAtts value in the map is ordered by profitability.
// The function takes the first attestation from each value, orders these attestations by bit count
// and places them at the start of the resulting slice. It then takes the second attestation for each value,
// orders these attestations by bit count and appends them to the end.
// It continues this pattern until all attestations are processed.
func sortSlotAttestations(slotAtts map[primitives.CommitteeIndex]ProposerAtts) ProposerAtts {
	attCount := 0
	for _, committeeAtts := range slotAtts {
		attCount += len(committeeAtts)
	}

	sorted := make([]ethpb.Att, 0, attCount)

	processedCount := 0
	index := 0
	for processedCount < attCount {
		var atts []ethpb.Att

		for _, committeeAtts := range slotAtts {
			if len(committeeAtts) > index {
				atts = append(atts, committeeAtts[index])
			}
		}

		sort.Slice(atts, func(i, j int) bool {
			return atts[i].GetAggregationBits().Count() > atts[j].GetAggregationBits().Count()
		})
		sorted = append(sorted, atts...)

		processedCount += len(atts)
		index++
	}

	return sorted
}

// sortByProfitabilityUsingMaxCover orders attestations by highest aggregation bit count.
// Duplicate bits are counted only once, using max-cover algorithm.
func (a ProposerAtts) sortByProfitabilityUsingMaxCover_committeeAwarePacking() (ProposerAtts, error) {
	if len(a) < 2 {
		return a, nil
	}
	candidates := make([]*bitfield.Bitlist64, len(a))
	for i := 0; i < len(a); i++ {
		var err error
		candidates[i], err = a[i].GetAggregationBits().ToBitlist64()
		if err != nil {
			return nil, err
		}
	}
	selectedKeys, _, err := aggregation.MaxCover(candidates, len(candidates), true /* allowOverlaps */)
	if err != nil {
		log.WithError(err).Debug("MaxCover aggregation failed")
		return a, nil
	}
	selected := make(ProposerAtts, selectedKeys.Count())
	for i, key := range selectedKeys.BitIndices() {
		selected[i] = a[key]
	}
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].GetAggregationBits().Count() > selected[j].GetAggregationBits().Count()
	})
	return selected, nil
}

// Separate attestations by slot, as slot number takes higher precedence when sorting.
// Also separate by committee index because maxcover will prefer attestations for the same
// committee with disjoint bits over attestations for different committees with overlapping
// bits, even though same bits for different committees are separate votes.
func (a ProposerAtts) sortBySlotAndCommittee() (ProposerAtts, error) {
	type slotAtts struct {
		candidates map[primitives.CommitteeIndex]ProposerAtts
		selected   map[primitives.CommitteeIndex]ProposerAtts
	}

	var slots []primitives.Slot
	attsBySlot := map[primitives.Slot]*slotAtts{}
	for _, att := range a {
		slot := att.GetData().Slot
		ci := att.GetData().CommitteeIndex
		if _, ok := attsBySlot[slot]; !ok {
			attsBySlot[slot] = &slotAtts{}
			attsBySlot[slot].candidates = make(map[primitives.CommitteeIndex]ProposerAtts)
			slots = append(slots, slot)
		}
		attsBySlot[slot].candidates[ci] = append(attsBySlot[slot].candidates[ci], att)
	}

	var err error
	for _, sa := range attsBySlot {
		sa.selected = make(map[primitives.CommitteeIndex]ProposerAtts)
		for ci, committeeAtts := range sa.candidates {
			sa.selected[ci], err = committeeAtts.sortByProfitabilityUsingMaxCover_committeeAwarePacking()
			if err != nil {
				return nil, err
			}
		}
	}

	var sortedAtts ProposerAtts
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] > slots[j]
	})
	for _, slot := range slots {
		sortedAtts = append(sortedAtts, sortSlotAttestations(attsBySlot[slot].selected)...)
	}

	return sortedAtts, nil
}

// Dedup removes duplicate attestations (ones with the same bits set on).
// Important: not only exact duplicates are removed, but proper subsets are removed too
// (their known bits are redundant and are already contained in their supersets).
func (a ProposerAtts) Dedup() (ProposerAtts, error) {
	if len(a) < 2 {
		return a, nil
	}
	attsByDataRoot := make(map[attestation.Id][]ethpb.Att, len(a))
	for _, att := range a {
		id, err := attestation.NewId(att, attestation.Data)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create attestation ID")
		}
		attsByDataRoot[id] = append(attsByDataRoot[id], att)
	}

	uniqAtts := make([]ethpb.Att, 0, len(a))
	for _, atts := range attsByDataRoot {
		for i := 0; i < len(atts); i++ {
			a := atts[i]
			for j := i + 1; j < len(atts); j++ {
				b := atts[j]
				if c, err := a.GetAggregationBits().Contains(b.GetAggregationBits()); err != nil {
					return nil, err
				} else if c {
					// a contains b, b is redundant.
					atts[j] = atts[len(atts)-1]
					atts[len(atts)-1] = nil
					atts = atts[:len(atts)-1]
					j--
				} else if c, err := b.GetAggregationBits().Contains(a.GetAggregationBits()); err != nil {
					return nil, err
				} else if c {
					// b contains a, a is redundant.
					atts[i] = atts[len(atts)-1]
					atts[len(atts)-1] = nil
					atts = atts[:len(atts)-1]
					i--
					break
				}
			}
		}
		uniqAtts = append(uniqAtts, atts...)
	}

	return uniqAtts, nil
}
