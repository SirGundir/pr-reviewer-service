package usecase

import (
	"context"
	"fmt"
	"pr-reviewer-service/internal/entity"
	"pr-reviewer-service/internal/usecase/repo"
	"time"
)

type PullRequestUseCase struct {
	prRepo   repo.PullRequestRepo
	userRepo repo.UserRepo
	selector *ReviewerSelector
}

func NewPullRequestUseCase(prr repo.PullRequestRepo, ur repo.UserRepo, rs *ReviewerSelector) *PullRequestUseCase {
	return &PullRequestUseCase{
		prRepo:   prr,
		userRepo: ur,
		selector: rs,
	}
}

func (uc *PullRequestUseCase) CreatePR(ctx context.Context, prID, prName, authorID string) (entity.PullRequest, error) {
	exists, err := uc.prRepo.Exists(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - CreatePR - uc.prRepo.Exists: %w", err)
	}
	if exists {
		return entity.PullRequest{}, entity.ErrPRAlreadyExists
	}

	author, err := uc.userRepo.GetByID(ctx, authorID)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - CreatePR - uc.userRepo.GetByID: %w", err)
	}

	teamMembers, err := uc.userRepo.GetByTeam(ctx, author.TeamName)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - CreatePR - uc.userRepo.GetByTeam: %w", err)
	}

	reviewers := uc.selector.SelectReviewers(teamMembers, authorID)

	pr := entity.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            entity.StatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         time.Now(),
	}

	if err := uc.prRepo.Create(ctx, pr); err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - CreatePR - uc.prRepo.Create: %w", err)
	}

	return pr, nil
}

func (uc *PullRequestUseCase) MergePR(ctx context.Context, prID string) (entity.PullRequest, error) {
	pr, err := uc.prRepo.GetByID(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - MergePR - uc.prRepo.GetByID: %w", err)
	}

	pr.Merge()

	if err := uc.prRepo.Update(ctx, pr); err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - MergePR - uc.prRepo.Update: %w", err)
	}

	return pr, nil
}

func (uc *PullRequestUseCase) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (entity.PullRequest, string, error) {
	pr, err := uc.prRepo.GetByID(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, "", fmt.Errorf("PullRequestUseCase - ReassignReviewer - uc.prRepo.GetByID: %w", err)
	}

	if pr.IsMerged() {
		return entity.PullRequest{}, "", entity.ErrPRAlreadyMerged
	}

	if !pr.HasReviewer(oldReviewerID) {
		return entity.PullRequest{}, "", entity.ErrReviewerNotAssigned
	}

	oldReviewer, err := uc.userRepo.GetByID(ctx, oldReviewerID)
	if err != nil {
		return entity.PullRequest{}, "", fmt.Errorf("PullRequestUseCase - ReassignReviewer - uc.userRepo.GetByID: %w", err)
	}

	teamMembers, err := uc.userRepo.GetByTeam(ctx, oldReviewer.TeamName)
	if err != nil {
		return entity.PullRequest{}, "", fmt.Errorf("PullRequestUseCase - ReassignReviewer - uc.userRepo.GetByTeam: %w", err)
	}

	newReviewerID, err := uc.selector.FindReplacement(teamMembers, pr.AuthorID, pr.AssignedReviewers)
	if err != nil {
		return entity.PullRequest{}, "", err
	}

	// Replace reviewer
	for i, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldReviewerID {
			pr.AssignedReviewers[i] = newReviewerID
			break
		}
	}

	if err := uc.prRepo.Update(ctx, pr); err != nil {
		return entity.PullRequest{}, "", fmt.Errorf("PullRequestUseCase - ReassignReviewer - uc.prRepo.Update: %w", err)
	}

	return pr, newReviewerID, nil
}
