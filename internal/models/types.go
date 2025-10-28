package models

// internal/models/types.go

type PeriodStats struct {
	Period     string `json:"period"`
	TotalPR    int    `json:"total_pr"`
	AIPR       int    `json:"ai_pr"`
	PRRate     string `json:"pr_rate"`
	TotalLines int    `json:"total_lines"`
	AILines    int    `json:"ai_lines"`
	CodeRate   string `json:"code_rate"`
}
