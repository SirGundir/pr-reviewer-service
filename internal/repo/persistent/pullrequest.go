package persistent

import (
	"context"
	"fmt"
	"pr-reviewer-service/internal/entity"
	"pr-reviewer-service/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type PullRequestRepo struct {
	*postgres.Postgres
}

func NewPullRequestRepo(pg *postgres.Postgres) *PullRequestRepo {
	return &PullRequestRepo{pg}
}

func (r *PullRequestRepo) Create(ctx context.Context, pr entity.PullRequest) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("PullRequestRepo - Create - r.Pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	sql, args, err := r.Builder.
		Insert("pull_requests").
		Columns("pull_request_id", "pull_request_name", "author_id", "status", "created_at").
		Values(pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, pr.CreatedAt).
		ToSql()

	if err != nil {
		return fmt.Errorf("PullRequestRepo - Create - r.Builder: %w", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PullRequestRepo - Create - tx.Exec: %w", err)
	}

	for _, reviewerID := range pr.AssignedReviewers {
		sql, args, err := r.Builder.
			Insert("pr_reviewers").
			Columns("pull_request_id", "reviewer_id").
			Values(pr.PullRequestID, reviewerID).
			ToSql()

		if err != nil {
			return fmt.Errorf("PullRequestRepo - Create - r.Builder (reviewers): %w", err)
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("PullRequestRepo - Create - tx.Exec (reviewers): %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PullRequestRepo) GetByID(ctx context.Context, prID string) (entity.PullRequest, error) {
	sql, args, err := r.Builder.
		Select("pull_request_id", "pull_request_name", "author_id", "status", "created_at", "merged_at").
		From("pull_requests").
		Where("pull_request_id = ?", prID).
		ToSql()

	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestRepo - GetByID - r.Builder: %w", err)
	}

	var pr entity.PullRequest
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)

	if err == pgx.ErrNoRows {
		return entity.PullRequest{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestRepo - GetByID - r.Pool.QueryRow: %w", err)
	}

	// Get reviewers
	reviewerSQL, reviewerArgs, err := r.Builder.
		Select("reviewer_id").
		From("pr_reviewers").
		Where("pull_request_id = ?", prID).
		ToSql()

	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestRepo - GetByID - r.Builder (reviewers): %w", err)
	}

	rows, err := r.Pool.Query(ctx, reviewerSQL, reviewerArgs...)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestRepo - GetByID - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return entity.PullRequest{}, fmt.Errorf("PullRequestRepo - GetByID - rows.Scan: %w", err)
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return pr, nil
}

func (r *PullRequestRepo) Update(ctx context.Context, pr entity.PullRequest) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("PullRequestRepo - Update - r.Pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	sql, args, err := r.Builder.
		Update("pull_requests").
		Set("status", pr.Status).
		Set("merged_at", pr.MergedAt).
		Where("pull_request_id = ?", pr.PullRequestID).
		ToSql()

	if err != nil {
		return fmt.Errorf("PullRequestRepo - Update - r.Builder: %w", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PullRequestRepo - Update - tx.Exec: %w", err)
	}

	// Delete old reviewers
	deleteSQL, deleteArgs, err := r.Builder.
		Delete("pr_reviewers").
		Where("pull_request_id = ?", pr.PullRequestID).
		ToSql()

	if err != nil {
		return fmt.Errorf("PullRequestRepo - Update - r.Builder (delete): %w", err)
	}

	_, err = tx.Exec(ctx, deleteSQL, deleteArgs...)
	if err != nil {
		return fmt.Errorf("PullRequestRepo - Update - tx.Exec (delete): %w", err)
	}

	// Insert new reviewers
	for _, reviewerID := range pr.AssignedReviewers {
		insertSQL, insertArgs, err := r.Builder.
			Insert("pr_reviewers").
			Columns("pull_request_id", "reviewer_id").
			Values(pr.PullRequestID, reviewerID).
			ToSql()

		if err != nil {
			return fmt.Errorf("PullRequestRepo - Update - r.Builder (insert): %w", err)
		}

		_, err = tx.Exec(ctx, insertSQL, insertArgs...)
		if err != nil {
			return fmt.Errorf("PullRequestRepo - Update - tx.Exec (insert): %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PullRequestRepo) GetByReviewer(ctx context.Context, userID string) ([]entity.PullRequest, error) {
	sql, args, err := r.Builder.
		Select("p.pull_request_id", "p.pull_request_name", "p.author_id", "p.status").
		From("pull_requests p").
		Join("pr_reviewers pr ON p.pull_request_id = pr.pull_request_id").
		Where("pr.reviewer_id = ?", userID).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("PullRequestRepo - GetByReviewer - r.Builder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("PullRequestRepo - GetByReviewer - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	var prs []entity.PullRequest
	for rows.Next() {
		var pr entity.PullRequest
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("PullRequestRepo - GetByReviewer - rows.Scan: %w", err)
		}
		prs = append(prs, pr)
	}

	return prs, nil
}

func (r *PullRequestRepo) Exists(ctx context.Context, prID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`

	var exists bool
	err := r.Pool.QueryRow(ctx, query, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("PullRequestRepo - Exists - r.Pool.QueryRow: %w", err)
	}

	return exists, nil
}

func (r *PullRequestRepo) GetUserStats(ctx context.Context) ([]entity.UserStats, error) {
	query := `
		SELECT 
			u.user_id,
			u.username,
			COUNT(pr.pull_request_id) as total_assigned,
			COUNT(CASE WHEN p.status = 'OPEN' THEN 1 END) as open_assigned,
			COUNT(CASE WHEN p.status = 'MERGED' THEN 1 END) as merged_assigned
		FROM users u
		LEFT JOIN pr_reviewers pr ON u.user_id = pr.reviewer_id
		LEFT JOIN pull_requests p ON pr.pull_request_id = p.pull_request_id
		GROUP BY u.user_id, u.username
		ORDER BY total_assigned DESC
	`

	rows, err := r.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("PullRequestRepo - GetUserStats - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	var stats []entity.UserStats
	for rows.Next() {
		var stat entity.UserStats
		if err := rows.Scan(&stat.UserID, &stat.Username, &stat.TotalAssigned, &stat.OpenAssigned, &stat.MergedAssigned); err != nil {
			return nil, fmt.Errorf("PullRequestRepo - GetUserStats - rows.Scan: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (r *PullRequestRepo) GetPRStats(ctx context.Context) (*entity.PRStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_prs,
			COUNT(CASE WHEN status = 'OPEN' THEN 1 END) as open_prs,
			COUNT(CASE WHEN status = 'MERGED' THEN 1 END) as merged_prs,
			(SELECT COUNT(*) FROM pr_reviewers) as total_reviewers
		FROM pull_requests
	`

	var stats entity.PRStats
	err := r.Pool.QueryRow(ctx, query).Scan(
		&stats.TotalPRs,
		&stats.OpenPRs,
		&stats.MergedPRs,
		&stats.TotalReviewers,
	)

	if err != nil {
		return nil, fmt.Errorf("PullRequestRepo - GetPRStats - r.Pool.QueryRow: %w", err)
	}

	return &stats, nil
}

func (r *PullRequestRepo) GetOpenPRsByTeam(ctx context.Context, teamName string) ([]entity.PullRequest, error) {
	query := `
		SELECT DISTINCT p.pull_request_id, p.pull_request_name, p.author_id, p.status, p.created_at, p.merged_at
		FROM pull_requests p
		JOIN pr_reviewers pr ON p.pull_request_id = pr.pull_request_id
		JOIN users u ON pr.reviewer_id = u.user_id
		WHERE u.team_name = $1 AND p.status = 'OPEN'
	`

	rows, err := r.Pool.Query(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("PullRequestRepo - GetOpenPRsByTeam: %w", err)
	}
	defer rows.Close()

	var prs []entity.PullRequest
	for rows.Next() {
		var pr entity.PullRequest
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	return prs, nil
}
