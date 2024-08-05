package slotstrategy

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

type NumberSlot int64

func (n NumberSlot) StrValue() string {
	return fmt.Sprintf("%d", n)
}

func (n NumberSlot) Compare(slot int64) int {
	log.WithFields(log.Fields{
		"compared slot": slot,
		"n":             n,
	}).Debug("compare slot")
	if int64(n) > slot {
		return 1
	}
	if int64(n) < slot {
		return -1
	}
	return 0
}
