package entity

type UserStats struct {
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	TotalAssigned  int    `json:"total_assigned"`
	OpenAssigned   int    `json:"open_assigned"`
	MergedAssigned int    `json:"merged_assigned"`
}

type PRStats struct {
	TotalPRs       int `json:"total_prs"`
	OpenPRs        int `json:"open_prs"`
	MergedPRs      int `json:"merged_prs"`
	TotalReviewers int `json:"total_reviewers"`
}
