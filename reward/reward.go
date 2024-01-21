package reward

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

type BeaconResponse struct {
	Data json.RawMessage `json:"data"`
}

func GetRewards(gwEndpoint string, output string) error {
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()
	client :=

	latestSlot, err := client.BlockNumber(context.Background())
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Epoch", "Validator Index", "Head", "Target", "Source", "Inclusion Delay", "Inactivity"})

	baseURL := fmt.Sprintf("%s/eth/v1/beacon/rewards/attestations/", addr)
	epochNumber := uint64(1)

	for epochNumber <= latestSlot {
		url := baseURL + strconv.FormatUint(epochNumber, 10)
		epochNumber++

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error fetching data for Epoch %d: %v\n", epochNumber, err)
			continue
		}
		defer resp.Body.Close()

		var response Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			log.Printf("Error decoding response for Epoch %d: %v\n", epochNumber, err)
			continue
		}

		for _, totalReward := range response.Data.TotalRewards {
			writer.Write([]string{strconv.FormatUint(epochNumber, 10), totalReward.ValidatorIndex, totalReward.Head, totalReward.Target, totalReward.Source, totalReward.InclusionDelay, totalReward.Inactivity})
		}
	}
	return err
}
