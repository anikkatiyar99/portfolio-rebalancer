package models

type Portfolio struct {
	UserID     string             `json:"user_id"`
	Allocation map[string]float64 `json:"allocation"` // Current user allocation in percentage terms
}

type UpdatedPortfolio struct {
	UserID        string             `json:"user_id"`
	NewAllocation map[string]float64 `json:"new_allocation"` // Updated user allocation from provider in percentage terms
}

type RebalanceTransaction struct {
	UserID           string  `json:"user_id"`
	Action           string  `json:"action"`            // can be BUY/SELL
	Asset            string  `json:"asset"`             // can be stocks, bonds, commodities, etc.
	RebalancePercent float64 `json:"rebalance_percent"` // how much to move
}
