package usecase

import (
	"context"
	"fmt"
	"pr-reviewer-service/internal/entity"
	"pr-reviewer-service/internal/usecase/repo"
)

type UserUseCase struct {
	userRepo repo.UserRepo
	prRepo   repo.PullRequestRepo
}

func NewUserUseCase(ur repo.UserRepo, prr repo.PullRequestRepo) *UserUseCase {
	return &UserUseCase{
		userRepo: ur,
		prRepo:   prr,
	}
}

func (uc *UserUseCase) SetIsActive(ctx context.Context, userID string, isActive bool) (entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserUseCase - SetIsActive - uc.userRepo.GetByID: %w", err)
	}

	if isActive {
		user.Activate()
	} else {
		user.Deactivate()
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return entity.User{}, fmt.Errorf("UserUseCase - SetIsActive - uc.userRepo.Update: %w", err)
	}

	return user, nil
}

func (uc *UserUseCase) GetReviews(ctx context.Context, userID string) ([]entity.PullRequest, error) {
	prs, err := uc.prRepo.GetByReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - GetReviews - uc.prRepo.GetByReviewer: %w", err)
	}
	return prs, nil
}
