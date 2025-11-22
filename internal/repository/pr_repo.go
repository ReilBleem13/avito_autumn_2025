package repository

import (
	"ReilBleem13/pull_requests_service/internal/domain"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PullRequestRepository struct {
	db *sqlx.DB
}

func NewPullRequestRepository(db *sqlx.DB) *PullRequestRepository {
	return &PullRequestRepository{
		db: db,
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

func (p *PullRequestRepository) Merge(ctx context.Context, prID string) error {
	tx, err := p.db.BeginTxx(ctx, &sql.TxOptions{})
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

	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM pull_requests
			WHERE pull_request_id = $1
		)
	`

	var exists bool
	if err := tx.GetContext(ctx, &exists, checkQuery, prID); err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("pull request is empty: %w", domain.ErrNotFound())
	}

	updateQuery := `
		UPDATE pull_requests 
		SET status = 'MERGED', 
			merged_at = NOW()
		WHERE pull_request_id = $1 AND status = 'OPEN'
	`

	_, err = tx.ExecContext(ctx, updateQuery, prID)
	if err != nil {
		return err
	}
	return nil
}

func (p *PullRequestRepository) GetPullRequest(ctx context.Context, prID string) (*domain.PullRequest, error) {
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

	getPRQuery := `
		SELECT 
			pull_request_id,
			pull_request_name,
			author_id,
			status,
			merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	if err := tx.GetContext(ctx, &pullRequest, getPRQuery, prID); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pull_request is not exits: %w", domain.ErrNotFound())
		}
		return nil, err
	}

	getQuery := `
		SELECT prr.user_id 
		FROM pull_request_reviewers prr
		JOIN pull_requests pr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.pull_request_id = $1 
	`
	// убрать проверку на pr.author_id (AND prr.user_id != pr.author_id)
	var users []string
	if err := tx.SelectContext(ctx, &users, getQuery, prID); err != nil {
		return nil, err
	}

	pullRequest.AssignedReviewers = users
	return &pullRequest, nil
}

func (p *PullRequestRepository) GetPullRequestByID(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	getQuery := `
		SELECT 
			pr.pull_request_id,
			pr.pull_request_name,
			pr.author_id,
			pr.status
		FROM pull_requests pr
		JOIN pull_request_reviewers prr ON prr.pull_request_id = pr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at
	`

	var pullRequests []domain.PullRequestShort
	if err := p.db.SelectContext(ctx, &pullRequests, getQuery, userID); err != nil {
		return nil, err
	}

	if len(pullRequests) == 0 {
		return nil, domain.ErrNotFound()
	}
	return pullRequests, nil
}

func (p *PullRequestRepository) ReAssign(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	reassignQuery := `
		UPDATE pull_request_reviewers 
		SET user_id = $3
		WHERE pull_request_id = $1 AND user_id = $2
	`

	res, err := p.db.ExecContext(ctx, reassignQuery, prID, oldReviewerID, newReviewerID)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return fmt.Errorf("pull_request was not found: %w", domain.ErrNotFound())
	}
	return nil
}
