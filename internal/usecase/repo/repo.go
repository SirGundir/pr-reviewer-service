package repo

import (
	"context"
	"pr-reviewer-service/internal/entity"
)

type (
	UserRepo interface {
		Create(ctx context.Context, user entity.User) error
		GetByID(ctx context.Context, userID string) (entity.User, error)
		Update(ctx context.Context, user entity.User) error
		GetByTeam(ctx context.Context, teamName string) ([]entity.User, error)
	}

	TeamRepo interface {
		Create(ctx context.Context, team entity.Team) error
		GetByName(ctx context.Context, teamName string) (entity.Team, error)
		Exists(ctx context.Context, teamName string) (bool, error)
	}

	PullRequestRepo interface {
		Create(ctx context.Context, pr entity.PullRequest) error
		GetByID(ctx context.Context, prID string) (entity.PullRequest, error)
		Update(ctx context.Context, pr entity.PullRequest) error
		GetByReviewer(ctx context.Context, userID string) ([]entity.PullRequest, error)
		Exists(ctx context.Context, prID string) (bool, error)
	}
)
