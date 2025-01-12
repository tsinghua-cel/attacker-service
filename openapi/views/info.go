package views

// That ReorgCount > 0
type StrategyWithReorgCount struct {
	ReorgCount      string `json:"reorg_count"`
	StrategyId      string `json:"strategy_id"`
	StrategyContent string `json:"strategy_content"`
}

// That HonestLoseRateAvg > 0
type StrategyWithHonestLose struct {
	HonestLoseRateAvg string `json:"honest_lose_rate_avg"`
	StrategyId        string `json:"strategy_id"`
	StrategyContent   string `json:"strategy_content"`
}

// That HonestLoseAvg > MaliciousLoseAvg
type StrategyWithGreatHonestLose struct {
	HonestLoseRateAvg    string `json:"honest_lose"`
	MaliciousLoseRateAvg string `json:"malicious_lose"`
	Ratio                string `json:"ratio"`
	StrategyId           string `json:"strategy_id"`
	StrategyContent      string `json:"strategy_content"`
}

type DashboardInfo struct {
	CurSlot           string `json:"cur_slot"`
	StrategyCount     string `json:"strategy_count"`
	LatestBlockHeight string `json:"latest_block_height"`
}
