package repository

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"ReilBleem13/pull_requests_service/internal/repository/database"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PullRequestRepository struct {
	db *sqlx.DB
}

func NewPullRequestRepository(db *database.PostgresDB) *PullRequestRepository {
	return &PullRequestRepository{
		db: db.Client(),
	}
}

func (p *PullRequestRepository) Create(ctx context.Context, prID, prName, authorID string, assignedUsers []string) error {
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	createPRQuery := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id)
		VALUES ($1, $2, $3)
	`
	_, err = tx.ExecContext(ctx, createPRQuery, prID, prName, authorID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.ErrPRExists()
		}
		return err
	}

	insertReviewerQuery := `
		INSERT INTO pull_request_reviewers (pull_request_id, user_id)
		VALUES ($1, $2)
	`
	for _, assignedUser := range assignedUsers {
		_, err = tx.ExecContext(ctx, insertReviewerQuery, prID, assignedUser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PullRequestRepository) GetActiveTeamMembers(ctx context.Context, teamName, authorID string) ([]domain.User, error) {
	getQuery := `
		SELECT u.user_id, u.username, u.is_active 
		FROM team_members tm
		JOIN users u ON u.user_id = tm.user_id
		WHERE tm.team_name = $1 
			AND u.is_active = true 
			AND u.user_id != $2
	`

	var users []domain.User
	if err := p.db.SelectContext(ctx, &users, getQuery, teamName, authorID); err != nil {
		return nil, err
	}
	return users, nil
}

func (p *PullRequestRepository) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
	tx, err := p.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var pullRequest domain.PullRequest
	getQuery := `
		SELECT
			pull_request_id,
			pull_request_name,
			author_id,
			status,
			merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	if err := tx.GetContext(ctx, &pullRequest, getQuery, prID); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pull_request is not exits: %w", sql.ErrNoRows)
		}
		return nil, err
	}

	updateQuery := `
		UPDATE pull_requests 
		SET status = 'MERGED', 
			merged_at = NOW()
		WHERE pull_request_id = $1 AND status = 'OPEN'
		RETURNING status, merged_at
	`

	err = tx.QueryRowContext(ctx, updateQuery, prID).Scan(
		&pullRequest.Status, &pullRequest.MergedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &pullRequest, nil
}

func (p *PullRequestRepository) GetAssignedReviewers(ctx context.Context, prID string) ([]string, error) {
	getQuery := `
		SELECT prr.user_id 
		FROM pull_request_reviewers prr
		JOIN pull_requests pr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.pull_request_id = $1 AND prr.user_id != pr.author_id
	`

	var users []string
	if err := p.db.SelectContext(ctx, &users, getQuery, prID); err != nil {
		return nil, err
	}
	return users, nil
}

func (p *PullRequestRepository) ReAssign() {}
