package reward

import (
	"encoding/csv"
	"github.com/tsinghua-cel/attacker-service/beaconapi"
	"os"
	"strconv"
)

func GetRewards(gwEndpoint string, output string) error {
	bakfile := output + ".bak"
	file, err := os.Create(bakfile)
	if err != nil {
		return err
	}
	succeed := false
	defer func() {
		file.Close()
		if succeed {
			os.Rename(bakfile, output)
		}
	}()
	client := beaconapi.NewBeaconGwClient(gwEndpoint)

	slots_per_epoch, err := client.GetIntConfig(beaconapi.SLOTS_PER_EPOCH)
	if err != nil {
		// todo: add log
		return err
	}
	latestHeader, err := client.GetLatestBeaconHeader()
	if err != nil {
		return err
	}

	latestSlot, _ := strconv.ParseInt(latestHeader.Header.Message.Slot, 10, 64)
	latestEpoch := latestSlot / int64(slots_per_epoch)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Epoch", "Validator Index", "Head", "Target", "Source", "Inclusion Delay", "Inactivity"})

	epochNumber := int64(0)

	for epochNumber <= (latestEpoch - 2) {
		info, err := client.GetAllValReward(int(epochNumber))
		if err != nil {
			return err
		}
		for _, totalReward := range info.TotalRewards {
			inclusionDelay := int64(0)
			if totalReward.InclusionDelay != nil {
				inclusionDelay = int64(*totalReward.InclusionDelay)
			}
			writer.Write([]string{
				strconv.FormatInt(epochNumber, 10),
				strconv.FormatInt(int64(totalReward.ValidatorIndex), 10),
				strconv.FormatInt(int64(totalReward.Head), 10),
				strconv.FormatInt(int64(totalReward.Target), 10),
				strconv.FormatInt(int64(totalReward.Source), 10),
				strconv.FormatInt(inclusionDelay, 10),
				strconv.FormatInt(int64(totalReward.Inactivity), 10),
			})
		}

		epochNumber++

	}
	succeed = true
	return err
}
