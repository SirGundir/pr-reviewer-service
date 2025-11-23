package persistent

import (
	"context"
	"fmt"
	"pr-reviewer-service/internal/entity"
	"pr-reviewer-service/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type TeamRepo struct {
	*postgres.Postgres
	userRepo *UserRepo
}

func NewTeamRepo(pg *postgres.Postgres, ur *UserRepo) *TeamRepo {
	return &TeamRepo{
		Postgres: pg,
		userRepo: ur,
	}
}

func (r *TeamRepo) Create(ctx context.Context, team entity.Team) error {
	sql, args, err := r.Builder.
		Insert("teams").
		Columns("team_name", "created_at").
		Values(team.TeamName, team.CreatedAt).
		ToSql()

	if err != nil {
		return fmt.Errorf("TeamRepo - Create - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("TeamRepo - Create - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *TeamRepo) GetByName(ctx context.Context, teamName string) (entity.Team, error) {
	sql, args, err := r.Builder.
		Select("team_name", "created_at").
		From("teams").
		Where("team_name = ?", teamName).
		ToSql()

	if err != nil {
		return entity.Team{}, fmt.Errorf("TeamRepo - GetByName - r.Builder: %w", err)
	}

	var team entity.Team
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&team.TeamName, &team.CreatedAt)

	if err == pgx.ErrNoRows {
		return entity.Team{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Team{}, fmt.Errorf("TeamRepo - GetByName - r.Pool.QueryRow: %w", err)
	}

	members, err := r.userRepo.GetByTeam(ctx, teamName)
	if err != nil {
		return entity.Team{}, fmt.Errorf("TeamRepo - GetByName - r.userRepo.GetByTeam: %w", err)
	}

	team.Members = members
	return team, nil
}

func (r *TeamRepo) Exists(ctx context.Context, teamName string) (bool, error) {
	sql, args, err := r.Builder.
		Select("EXISTS(SELECT 1 FROM teams WHERE team_name = ?)").
		From("teams").
		Where("team_name = ?", teamName).
		ToSql()

	if err != nil {
		return false, fmt.Errorf("TeamRepo - Exists - r.Builder: %w", err)
	}

	var exists bool
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("TeamRepo - Exists - r.Pool.QueryRow: %w", err)
	}

	return exists, nil
}
