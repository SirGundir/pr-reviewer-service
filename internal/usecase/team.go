package usecase

import (
	"context"
	"fmt"
	"pr-reviewer-service/internal/entity"
	"pr-reviewer-service/internal/usecase/repo"
	"time"
)

type TeamUseCase struct {
	teamRepo repo.TeamRepo
	userRepo repo.UserRepo
}

func NewTeamUseCase(tr repo.TeamRepo, ur repo.UserRepo) *TeamUseCase {
	return &TeamUseCase{
		teamRepo: tr,
		userRepo: ur,
	}
}

func (uc *TeamUseCase) CreateTeam(ctx context.Context, teamName string, members []entity.User) (entity.Team, error) {
	exists, err := uc.teamRepo.Exists(ctx, teamName)
	if err != nil {
		return entity.Team{}, fmt.Errorf("TeamUseCase - CreateTeam - uc.teamRepo.Exists: %w", err)
	}
	if exists {
		return entity.Team{}, entity.ErrTeamAlreadyExists
	}

	team := entity.Team{
		TeamName:  teamName,
		Members:   members,
		CreatedAt: time.Now(),
	}

	if err := uc.teamRepo.Create(ctx, team); err != nil {
		return entity.Team{}, fmt.Errorf("TeamUseCase - CreateTeam - uc.teamRepo.Create: %w", err)
	}

	for _, member := range members {
		member.TeamName = teamName
		member.CreatedAt = time.Now()
		if err := uc.userRepo.Create(ctx, member); err != nil {
			return entity.Team{}, fmt.Errorf("TeamUseCase - CreateTeam - uc.userRepo.Create: %w", err)
		}
	}

	return team, nil
}

func (uc *TeamUseCase) GetTeam(ctx context.Context, teamName string) (entity.Team, error) {
	team, err := uc.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return entity.Team{}, fmt.Errorf("TeamUseCase - GetTeam - uc.teamRepo.GetByName: %w", err)
	}
	return team, nil
}

func (uc *TeamUseCase) DeactivateTeamAndReassign(ctx context.Context, teamName string, prRepo repo.PullRequestRepo, selector *ReviewerSelector) error {
	openPRs, err := prRepo.GetOpenPRsByTeam(ctx, teamName)
	if err != nil {
		return fmt.Errorf("TeamUseCase - DeactivateTeamAndReassign - GetOpenPRs: %w", err)
	}

	if err := uc.userRepo.DeactivateTeam(ctx, teamName); err != nil {
		return fmt.Errorf("TeamUseCase - DeactivateTeamAndReassign - DeactivateTeam: %w", err)
	}

	for _, pr := range openPRs {
		fullPR, err := prRepo.GetByID(ctx, pr.PullRequestID)
		if err != nil {
			continue
		}

		author, err := uc.userRepo.GetByID(ctx, fullPR.AuthorID)
		if err != nil {
			continue
		}

		teamMembers, err := uc.userRepo.GetByTeam(ctx, author.TeamName)
		if err != nil {
			continue
		}

		newReviewers := selector.SelectReviewers(teamMembers, fullPR.AuthorID)
		fullPR.AssignedReviewers = newReviewers

		prRepo.Update(ctx, fullPR)
	}

	return nil
}
