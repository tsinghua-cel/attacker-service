package reward

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"
)

type TotalReward struct {
	ValidatorIndex string `json:"validator_index"`
	Head           string `json:"head"`
	Target         string `json:"target"`
	Source         string `json:"source"`
	InclusionDelay string `json:"inclusion_delay"`
	Inactivity     string `json:"inactivity"`
}

type RewardInfo struct {
	TotalRewards []TotalReward `json:"total_rewards"`
}

type BeaconHeaderInfo struct {
	Header struct {
		Message struct {
			Slot          string `json:"slot"`
			ProposerIndex string `json:"proposer_index"`
			ParentRoot    string `json:"parent_root"`
			StateRoot     string `json:"state_root"`
			BodyRoot      string `json:"body_root"`
		} `json:"message"`
		Signature string `json:"signature"`
	} `json:"header"`
	Root      string `json:"root"`
	Canonical bool   `json:"canonical"`
}

type BeaconResponse struct {
	Data json.RawMessage `json:"data"`
}

func GetRewards(gwEndpoint string, output string) error {
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()
	client := NewBeaconGwClient(gwEndpoint)

	slots_per_epoch, err := client.GetIntConfig(SLOTS_PER_EPOCH)
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
		totalRewards, err := client.GetAllValReward(int(epochNumber))
		if err != nil {
			return err
		}
		for _, totalReward := range totalRewards {
			writer.Write([]string{strconv.FormatInt(epochNumber, 10), totalReward.ValidatorIndex, totalReward.Head, totalReward.Target, totalReward.Source, totalReward.InclusionDelay, totalReward.Inactivity})
		}

		epochNumber++

	}
	return err
}
