package collection

import (
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"github.com/tsinghua-cel/attacker-service/dbmodel"
	"strconv"
	"time"
)

func ScheduleSlotUpdate(interval time.Duration, url string) {
	client := beaconapi.NewBeaconGwClient(url)
	tc := time.NewTicker(interval)
	defer tc.Stop()
	for {
		select {
		case <-tc.C:
			log.Info("ScheduleSlotUpdate")
			if err := UpdateProjectSlot(client); err != nil {
				log.WithError(err).Error("ScheduleBlockDuty failed")
			}
		}
	}
}

func UpdateProjectSlot(client *beaconapi.BeaconGwClient) error {
	latestHeader, err := client.GetLatestBeaconHeader()
	if err != nil {
		return err
	}

	latestSlot, _ := strconv.ParseInt(latestHeader.Header.Message.Slot, 10, 64)
	dbmodel.UpdateProjectLatestSlot(latestSlot)
	return nil
}
