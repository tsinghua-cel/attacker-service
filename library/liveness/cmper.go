package liveness

type ProposerListCmper interface {
	BetterThan(a, b BestMaskDutyInfo) bool
}

type AttackerCountCmper struct{}

func (a AttackerCountCmper) BetterThan(infoA, infoB BestMaskDutyInfo) bool {
	return infoA.AttackersCount >= infoB.AttackersCount
}

type AttackerCountAndFirstAttackerCmper struct{}

func (a AttackerCountAndFirstAttackerCmper) BetterThan(infoA, infoB BestMaskDutyInfo) bool {
	if infoA.FirstIsAttack && !infoB.FirstIsAttack {
		return true
	} else if !infoA.FirstIsAttack && infoB.FirstIsAttack {
		return false
	}

	// prefer the one with more attackers.
	return infoA.AttackersCount >= infoB.AttackersCount
}
