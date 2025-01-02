package types

import "strconv"

type AttackerInc interface {
	UpdateStrategy(strategy Strategy) error
	GetStrategyFeedback(uid string) (FeedBackInfo, error)
	GetChainBaseInfo() ChainBaseInfo
	GetCurSlot() int64
	GetEpochDuties(epoch int64) ([]ProposerDuty, error)
}

type LibraryParams struct {
	Attacker          AttackerInc
	MaxValidatorIndex int
	MinValidatorIndex int
}

func (p LibraryParams) GetLatestHackerSlot(duties []ProposerDuty) int {
	latest, _ := strconv.Atoi(duties[0].Slot)
	for _, duty := range duties {
		idx, _ := strconv.Atoi(duty.ValidatorIndex)
		slot, _ := strconv.Atoi(duty.Slot)
		if !p.IsHackValidator(idx) {
			continue
		}
		if slot > latest {
			latest = slot
		}
	}
	return latest
}

func (p LibraryParams) IsHackValidator(valIdx int) bool {
	return valIdx >= p.MinValidatorIndex && valIdx <= p.MaxValidatorIndex
}

func (p LibraryParams) FillterHackerDuties(duties []ProposerDuty) []ProposerDuty {
	res := make([]ProposerDuty, 0)
	for _, duty := range duties {
		idx, _ := strconv.Atoi(duty.ValidatorIndex)
		if p.IsHackValidator(idx) {
			res = append(res, duty)
		}
	}
	return res
}
