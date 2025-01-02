package confuse

import (
	"github.com/tsinghua-cel/attacker-service/types"
)

func CheckDuties(param types.LibraryParams, duties []types.ProposerDuty) ([]types.ProposerDuty, bool) {
	result := param.FillterHackerDuties(duties)
	return result, len(result) > 0
}
