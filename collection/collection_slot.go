package collection

import (
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"github.com/tsinghua-cel/attacker-service/dbmodel"
	"strconv"
)

func UpdateProjectSlot(client *beaconapi.BeaconGwClient) error {
	latestHeader, err := client.GetLatestBeaconHeader()
	if err != nil {
		return err
	}

	latestSlot, _ := strconv.ParseInt(latestHeader.Header.Message.Slot, 10, 64)
	dbmodel.UpdateProjectLatestSlot(latestSlot)
	return nil
}
