package models

type Stats struct {
	NumberOfTotalTransactions int64 `json:"total_transactions_count"`
	NumberOfLastWeekTransactions int64 `json:"last_week_transactions_count"`
	NumberOf24HoursTransactions int64 `json:"24hours_transactions_count"`
	// NumberOfBlocks       int64 `json:"total_blocks_count"`
}
type Richlist struct {
	TotalSupply       string     `json:"total_supply"`
	CirculatingSupply string     `json:"circulating_supply"`
	Rankings          []*Address `json:"rankings"`
}
