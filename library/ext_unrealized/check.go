package ext_unrealized

import (
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

func CheckDuties(param types.LibraryParams, duties []types.ProposerDuty) ([]interface{}, bool) {
	result := make([]interface{}, 0)

	tmpsub := make([]types.ProposerDuty, 0)

	firstproposerindex, _ := strconv.Atoi(duties[0].ValidatorIndex)
	if !param.IsHackValidator(firstproposerindex) {
		return nil, false
	}
	duty := duties[0]
	tmpsub = append(tmpsub, duty)
	result = append(result, tmpsub)
	return result, true
}
