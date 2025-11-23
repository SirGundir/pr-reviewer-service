package usecase

import (
	"math/rand"
	"pr-reviewer-service/internal/entity"
)

type ReviewerSelector struct{}

func NewReviewerSelector() *ReviewerSelector {
	return &ReviewerSelector{}
}

func (rs *ReviewerSelector) SelectReviewers(teamMembers []entity.User, authorID string) []string {
	candidates := rs.getCandidates(teamMembers, authorID, []string{})
	return rs.randomPick(candidates, 2)
}

func (rs *ReviewerSelector) FindReplacement(teamMembers []entity.User, authorID string, currentReviewers []string) (string, error) {
	candidates := rs.getCandidates(teamMembers, authorID, currentReviewers)

	if len(candidates) == 0 {
		return "", entity.ErrNoCandidates
	}

	selected := rs.randomPick(candidates, 1)
	if len(selected) == 0 {
		return "", entity.ErrNoCandidates
	}

	return selected[0], nil
}

func (rs *ReviewerSelector) getCandidates(teamMembers []entity.User, authorID string, exclude []string) []string {
	excludeMap := make(map[string]bool)
	excludeMap[authorID] = true
	for _, userID := range exclude {
		excludeMap[userID] = true
	}

	var candidates []string
	for _, member := range teamMembers {
		if member.IsActive && !excludeMap[member.UserID] {
			candidates = append(candidates, member.UserID)
		}
	}

	return candidates
}

func (rs *ReviewerSelector) randomPick(candidates []string, count int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	if count > len(candidates) {
		count = len(candidates)
	}

	shuffled := make([]string, len(candidates))
	copy(shuffled, candidates)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:count]
}
