package persistent

import (
	"context"
	"fmt"
	"pr-reviewer-service/internal/entity"
	"pr-reviewer-service/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	*postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (r *UserRepo) Create(ctx context.Context, user entity.User) error {
	sql, args, err := r.Builder.
		Insert("users").
		Columns("user_id", "username", "team_name", "is_active", "created_at").
		Values(user.UserID, user.Username, user.TeamName, user.IsActive, user.CreatedAt).
		Suffix("ON CONFLICT (user_id) DO UPDATE SET username = ?, team_name = ?, is_active = ?",
			user.Username, user.TeamName, user.IsActive).
		ToSql()

	if err != nil {
		return fmt.Errorf("UserRepo - Create - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo - Create - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, userID string) (entity.User, error) {
	sql, args, err := r.Builder.
		Select("user_id", "username", "team_name", "is_active", "created_at").
		From("users").
		Where("user_id = ?", userID).
		ToSql()

	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepo - GetByID - r.Builder: %w", err)
	}

	var user entity.User
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return entity.User{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepo - GetByID - r.Pool.QueryRow: %w", err)
	}

	return user, nil
}

func (r *UserRepo) Update(ctx context.Context, user entity.User) error {
	sql, args, err := r.Builder.
		Update("users").
		Set("username", user.Username).
		Set("team_name", user.TeamName).
		Set("is_active", user.IsActive).
		Where("user_id = ?", user.UserID).
		ToSql()

	if err != nil {
		return fmt.Errorf("UserRepo - Update - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo - Update - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserRepo) GetByTeam(ctx context.Context, teamName string) ([]entity.User, error) {
	sql, args, err := r.Builder.
		Select("user_id", "username", "team_name", "is_active", "created_at").
		From("users").
		Where("team_name = ?", teamName).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetByTeam - r.Builder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetByTeam - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("UserRepo - GetByTeam - rows.Scan: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepo) DeactivateTeam(ctx context.Context, teamName string) error {
	sql, args, err := r.Builder.
		Update("users").
		Set("is_active", false).
		Where("team_name = ?", teamName).
		ToSql()

	if err != nil {
		return fmt.Errorf("UserRepo - DeactivateTeam - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo - DeactivateTeam - r.Pool.Exec: %w", err)
	}

	return nil
}
