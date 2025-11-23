package usecase

import (
	"context"
	"fmt"
	"pr-reviewer-service/internal/entity"
	"pr-reviewer-service/internal/usecase/repo"
)

type StatsUseCase struct {
	prRepo   repo.PullRequestRepo
	userRepo repo.UserRepo
}

func NewStatsUseCase(prRepo repo.PullRequestRepo, userRepo repo.UserRepo) *StatsUseCase {
	return &StatsUseCase{
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}

func (uc *StatsUseCase) GetUserStats(ctx context.Context) ([]entity.UserStats, error) {
	stats, err := uc.prRepo.GetUserStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("StatsUseCase - GetUserStats: %w", err)
	}
	return stats, nil
}

func (uc *StatsUseCase) GetPRStats(ctx context.Context) (*entity.PRStats, error) {
	stats, err := uc.prRepo.GetPRStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("StatsUseCase - GetPRStats: %w", err)
	}
	return stats, nil
}
