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
	sql, args, err := r.Builder.
		Select("EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = ?)").
		From("pull_requests").
		Where("pull_request_id = ?", prID).
		ToSql()

	if err != nil {
		return false, fmt.Errorf("PullRequestRepo - Exists - r.Builder: %w", err)
	}

	var exists bool
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("PullRequestRepo - Exists - r.Pool.QueryRow: %w", err)
	}

	return exists, nil
}
