package entity

import "errors"

var (
	ErrTeamAlreadyExists   = errors.New("team already exists")
	ErrPRAlreadyExists     = errors.New("pull request already exists")
	ErrPRAlreadyMerged     = errors.New("PR is already merged")
	ErrReviewerNotAssigned = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidates        = errors.New("no candidate reviewers available")
	ErrNotFound            = errors.New("not found")
	ErrMaxReviewers        = errors.New("maximum 2 reviewers allowed")
)
