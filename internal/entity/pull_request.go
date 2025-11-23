package entity

import "time"

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            PRStatus   `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"created_at"`
	MergedAt          *time.Time `json:"merged_at,omitempty"`
}

func (pr *PullRequest) IsMerged() bool {
	return pr.Status == StatusMerged
}

func (pr *PullRequest) HasReviewer(userID string) bool {
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == userID {
			return true
		}
	}
	return false
}

func (pr *PullRequest) Merge() {
	if pr.Status != StatusMerged {
		pr.Status = StatusMerged
		now := time.Now()
		pr.MergedAt = &now
	}
}
